# Bug: Buku Besar, Neraca Saldo, dan Neraca (mode period_id) tidak kumulatif lintas periode

**Severity**: Tinggi — laporan menampilkan saldo akun riil yang salah begitu ada lebih dari satu periode akuntansi. Baru kelihatan sekarang karena proses closing (Fase 2 `docs/plan/penyempurnaan-akuntansi.md`) baru pertama kali benar-benar berfungsi.

> **Status: Diperbaiki** — ketiga laporan kini kumulatif sampai `end_date` periode yang dipilih (carry-forward akun riil): `report_trial_balance.go` & `report_ledger.go` memakai `je.entry_date <= <end_date>`, `report_balance_sheet.go` menghapus `computePeriodBalances` dan selalu memakai `computeBalancesToDate` (period_id di-resolve ke end_date). Buku Besar kini punya `opening_balance` (kumulatif sebelum start_date) dan `closing_balance`. Bonus: normalisasi tanda normal-balance di Neraca (liability/equity credit-normal tidak lagi tampil negatif).

## Prinsip yang dilanggar

Akun riil (`asset`, `liability`, `equity` — Kas, Bank, Piutang, Utang, Modal, Saldo Laba) **kumulatif**: saldonya terus dibawa (carry-forward) lintas periode, tidak pernah direset. Hanya akun nominal (`revenue`, `expense`) yang direset ke nol tiap periode lewat jurnal closing (`docs/rules/akuntansi.md` §3.2). Laporan yang menghitung saldo akun riil **hanya** dari jurnal dalam satu `period_id` akan salah begitu ada transaksi/saldo dari periode sebelumnya.

## Lokasi

| File | Query | Masalah |
|---|---|---|
| `report_ledger.go:31-42` | `WHERE jel.account_id = $1 AND je.period_id = $2` | Saldo berjalan (`runningBalance`) mulai dari 0 di setiap periode — tidak ada saldo awal dibawa dari periode sebelumnya |
| `report_trial_balance.go:30-38` | `WHERE je.period_id = $1` | Cuma ada mode ini, tidak ada opsi kumulatif sama sekali |
| `report_balance_sheet.go:92-104` (`computePeriodBalances`) | `WHERE je.period_id = $1` | Salah — tapi file ini **juga punya** `computeBalancesToDate` (baris 106-122, `WHERE je.entry_date <= $1`) yang **benar** (kumulatif). `Execute` (baris 39-44) memilih salah satu berdasarkan query param mana yang dikirim — kalau caller kirim `period_id`, dapat versi yang salah. |
| `report_income_statement.go` | `WHERE je.period_id = $1` | **Bukan bug** — akun nominal (revenue/expense) memang harus direset tiap periode, period-scoping di sini benar. |

## Gejala konkret

Skenario reproduksi:
1. Buat Periode Akuntansi "Juli 2026", posting beberapa invoice/payment (Kas & Piutang bertambah).
2. Tutup periode Juli — jurnal `closing` terbuat dengan `period_id` = periode Juli, memindahkan saldo revenue/expense ke `3201`/`3200`.
3. Buka Periode Akuntansi "Agustus 2026".
4. Panggil `GET /admin/reports/ledger?account_id=<Kas>&period_id=<Agustus>` atau `/reports/trial-balance?period_id=<Agustus>` atau `/reports/balance-sheet?period_id=<Agustus>`.

Hasil: saldo Kas, Piutang, dan `3200 Saldo Laba` semuanya tampil seolah mulai dari nol lagi di Agustus — hilang total saldo yang seharusnya terbawa dari Juli (termasuk hasil closing-nya sendiri, karena jurnal `closing` itu `period_id`-nya Juli, bukan Agustus). Kalau periode Agustus sendiri belum banyak transaksi, angka yang tampil bisa sangat kecil/nol padahal saldo riil jauh lebih besar — inilah yang terasa "janggal".

## Dampak

- Bendahara/pengurus yang lihat Neraca atau Neraca Saldo periode berjalan akan melihat posisi keuangan yang jauh lebih kecil dari kenyataan begitu sudah lewat dari periode pertama.
- Buku Besar per akun tidak bisa dipakai untuk rekonsiliasi (mis. cocokkan saldo Kas ke buku fisik) karena "saldo akhir" yang ditampilkan bukan saldo riil.
- Neraca Saldo (`report_trial_balance.go`) tidak punya jalan keluar sama sekali (tidak ada parameter alternatif) — beda dengan Neraca yang setidaknya punya mode `as_of_date` yang benar.

## Saran Perbaikan

Prinsip: **jangan ubah kontrak API** (tetap terima `period_id` dari dropdown pilihan periode di UI, konsisten dengan laporan lain) — tapi di dalam query, ganti jadi kumulatif sampai `end_date` periode yang dipilih, alih-alih membatasi ke `period_id` itu saja. Pola yang benar sudah ada di `computeBalancesToDate` (`report_balance_sheet.go:106-122`), tinggal direplikasi.

1. **`report_trial_balance.go`**: ambil `period.EndDate` dulu (`periodRepo.FindByID`, sudah di-inject), lalu ganti query:
   ```sql
   WHERE je.entry_date <= $1 AND je.status = 'posted'
   ```
   (parameter `$1` = `period.EndDate`, bukan `period.ID`).
2. **`report_ledger.go`**: sama, ambil `period.EndDate` lewat `periodRepo` (perlu ditambahkan sebagai dependency — saat ini `ReportLedgerUseCase` cuma punya `accountRepo`, belum punya `periodRepo`). Ganti `AND je.period_id = $2` jadi `AND je.entry_date <= $2` (nilai = `period.EndDate`). **Tambahan penting**: laporan buku besar konvensinya menampilkan "Saldo Awal" sebagai baris pertama (saldo kumulatif sampai sehari sebelum `period.StartDate`), baru diikuti transaksi *dalam* periode itu saja sebagai baris detail, dan "Saldo Akhir" di akhir — bukan cuma dump semua transaksi dari awal waktu tercampur tanpa pemisah. Desain ulang response `LedgerResponse` untuk membedakan "saldo awal" vs "transaksi periode ini" kalau mau sesuai konvensi buku besar yang lazim; kalau ingin tetap sederhana, minimal tampilkan saldo awal sebagai satu baris terpisah sebelum baris transaksi periode ini dimulai.
3. **`report_balance_sheet.go`**: `computePeriodBalances` sebaiknya **dihapus**, dan `Execute` selalu pakai `computeBalancesToDate` — kalau `query.PeriodID` diisi, resolve dulu jadi `period.EndDate` lewat `periodRepo.FindByID` lalu panggil `computeBalancesToDate` dengan tanggal itu (perlu tambah `periodRepo` sebagai dependency, saat ini `ReportBalanceSheetUseCase` belum punya). Ini menyatukan dua mode yang sekarang bercabang jadi satu jalur yang benar, menghindari kelas bug "salah pilih mode" terulang.
4. **`report_income_statement.go`**: **tidak diubah** — sudah benar, tetap `period_id`-scoped.

Verifikasi: ulangi skenario reproduksi di atas setelah perbaikan — saldo Kas/Piutang/Saldo Laba di Agustus harus menunjukkan saldo kumulatif yang benar (termasuk carry-forward dari Juli dan hasil closing-nya), sementara Laporan Laba Rugi Agustus tetap mulai dari nol untuk revenue/expense.
