# Bug: 4 dari 6 laporan akuntansi query kolom `deleted_at` yang tidak ada di `journal_entries`/`journal_entry_lines`

**Severity**: Tinggi — bukan cuma 2 laporan seperti temuan awal, tapi **4 dari 6 laporan akuntansi** akan selalu gagal (error DB), bukan sekadar menampilkan data kosong.

> **Status: Diperbaiki** — klausa `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL` dihapus dari `report_ledger.go`, `report_trial_balance.go`, `report_income_statement.go`, dan `report_balance_sheet.go` (dua query), diganti `AND je.status = 'posted'`.

## Lokasi (dikonfirmasi lewat `grep -n "deleted_at" internal/modules/keuangan/application/query/report_*.go`)

| File | Baris | Query |
|---|---|---|
| `report_ledger.go` | 41-42 | `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL` |
| `report_trial_balance.go` | 37-38 | `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL` |
| `report_income_statement.go` | 43-44 | `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL` |
| `report_balance_sheet.go` | 100-101 dan 114-115 (dua query terpisah di file yang sama) | `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL` |

`report_summary.go` dan `report_outstanding.go` **tidak** kena bug ini — keduanya hanya memfilter `i.deleted_at` (tabel `invoices`, yang memang punya kolom itu).

Skema aktual (`migrations/*_create_keuangan_tables.up.sql`, tabel `journal_entries` & `journal_entry_lines`): **tidak ada kolom `deleted_at`** di kedua tabel ini. Jurnal tidak pernah soft-delete — pembatalan dilakukan lewat `status='cancelled'` (lihat `JournalEntry.Cancel()`), bukan penghapusan baris.

## Gejala

Memanggil salah satu dari empat endpoint berikut akan menghasilkan error dari Postgres (`column "deleted_at" does not exist`), diteruskan sebagai 500 ke klien:
- `GET /admin/reports/ledger`
- `GET /admin/reports/trial-balance`
- `GET /admin/reports/income-statement`
- `GET /admin/reports/balance-sheet`

Praktis **seluruh laporan yang berbasis jurnal tidak bisa dipakai sama sekali** dalam kondisi sekarang, terlepas dari isu auto-posting yang belum terpasang.

## Akar Masalah

Kemungkinan disalin dari pola query tabel lain di modul ini yang memang punya soft-delete (mis. `invoices`, `accounts`), tanpa menyesuaikan ke skema `journal_entries`/`journal_entry_lines` yang tidak memilikinya. Karena empat file ini isinya mirip (kemungkinan hasil copy-paste satu sama lain), begitu satu diperbaiki, cek tiga lainnya di daftar lokasi di atas — jangan hanya perbaiki satu lalu anggap selesai.

## Dampak

4 dari 6 laporan akuntansi (semua kecuali "Rekap Tagihan & Pembayaran" dan "Tunggakan per Santri", yang basisnya tabel `invoices` bukan jurnal) tidak berfungsi.

## Saran Perbaikan

Di keempat file, hapus klausa `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL` (dan variannya di `report_balance_sheet.go` yang punya dua kemunculan). Sebagai gantinya, tambahkan **`AND je.status = 'posted'`** di WHERE clause yang sama (lihat `docs/rules/akuntansi.md` §4.7) — supaya jurnal `cancelled` (dan `draft`, walau seharusnya tidak pernah tersisa sebagai draft di DB karena `Post()` selalu dipanggil sebelum `Save()`) tidak ikut terhitung di laporan manapun.
