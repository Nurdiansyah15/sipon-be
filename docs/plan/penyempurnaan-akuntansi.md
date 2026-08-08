# Rencana: Penyempurnaan Tagihan, Pembayaran, COA, Jurnal & Periode Akuntansi

## Context

Modul `keuangan` sudah punya billing (tagihan, pembayaran) dan akuntansi (COA, jurnal, periode) yang masing-masing berfungsi sendiri-sendiri, tapi **belum saling terhubung**: invoice yang diterbitkan dan pembayaran yang diverifikasi tidak pernah menghasilkan jurnal, walau service untuk itu (`AutoPostingService`) sudah ditulis lengkap sejak awal. Proses tutup buku (closing) periode akuntansi juga belum benar-benar memindahkan saldo — cuma flip status.

Dokumen ini adalah rencana untuk menyambungkan semuanya, dipecah ke tiga dokumen rujukan supaya masing-masing bisa berdiri sendiri dan gampang dirujuk saat implementasi:

- **[`docs/rules/akuntansi.md`](../rules/akuntansi.md)** — aturan bisnis: kapan jurnal dibuat, oleh siapa, aturan periode & closing, dan aturan integritas COA lainnya. Ini rujukan utama kalau ada pertanyaan "seharusnya begini atau begitu?".
- **[`docs/schemas/keuangan-akuntansi.md`](../schemas/keuangan-akuntansi.md)** — skema tabel saat ini + perubahan yang diusulkan (5 perubahan, lihat tabel ringkasan di dokumen itu).
- **[`docs/specs/keuangan-akuntansi-api.md`](../specs/keuangan-akuntansi-api.md)** — endpoint yang sudah ada + perubahan perilaku/endpoint baru.
- **[`docs/bugs/`](../bugs/)** — 7 bug konkret yang ditemukan saat riset untuk dokumen ini (lihat `docs/bugs/README.md` untuk daftar lengkap), sebagian besar justru **akar penyebab** kenapa penyambungan ini belum pernah kejadian atau akan langsung gagal begitu disambungkan tanpa diperbaiki dulu.

Lingkup sengaja dijaga sederhana — organisasi pengurus pondok pesantren/sekolah, bukan entitas komersial: tidak ada multi-currency, pajak otomatis, depresiasi otomatis, atau refund (lihat `docs/rules/akuntansi.md` §1.6).

## Kenapa urutan fase di bawah begini

Auto-posting tidak bisa asal disambung ke `create_invoice`/`verify_payment` sebelum bug-bug fondasinya (nomor jurnal, baris jurnal tidak tersimpan, laporan query kolom yang tidak ada) diperbaiki — begitu volume transaksi meningkat, ketiga bug itu akan langsung kelihatan (nomor jurnal manual sudah pasti tabrakan di transaksi kedua, laporan ledger/trial balance sudah pasti 500 sejak endpoint pertama kali dipanggil). Maka: **bereskan fondasi dulu, baru sambungkan.**

---

## FASE 0 — Perbaikan Bug Fondasi (harus selesai duluan)

Semua bug ini didokumentasikan lengkap (lokasi, gejala, saran perbaikan) di `docs/bugs/`. Ringkasan urutan kerja:

1. **`docs/bugs/akuntansi-journal-lines-tidak-tersimpan.md`** — `PostgresJournalRepository.Save` harus memanggil `SaveLines`. Sekalian tambahkan `transactor.WithTx` ke `CreateManualJournalUseCase` (belum ada sama sekali).
2. **`docs/bugs/akuntansi-manual-journal-nomor-selalu-1.md`** — tambah `NextJournalNumber` ke `JournalRepository`, pakai `finance_number_sequences` (`doc_type='journal'`), pola sama seperti `NextInvoiceNumber`/`NextPaymentNumber`.
3. **`docs/bugs/akuntansi-laporan-ledger-trial-balance-kolom-deleted-at.md`** — hapus filter `deleted_at` yang tidak ada, ganti jadi filter `status='posted'`.
4. **`docs/bugs/akuntansi-cancel-invoice-partial-payment.md`** — `Invoice.Cancel()` tolak kalau `PaidAmount > 0`, bukan cuma status `paid`.
5. **`docs/bugs/akuntansi-cancel-journal-tidak-cek-periode.md`** — `CancelJournalUseCase` tolak kalau periode jurnal tidak `open`.

**Verifikasi Fase 0**: `go build/vet/test`. Manual: buat 2 jurnal manual di bulan yang sama (nomor tidak boleh tabrakan, baris tersimpan & muncul di `GET /admin/journal-entries/:id`); panggil `GET /admin/reports/ledger` & `/reports/trial-balance` (tidak boleh 500); coba batalkan invoice `partial` (harus ditolak); coba batalkan jurnal di periode `closed` (harus ditolak).

---

## FASE 1 — Sambungkan Auto-Posting ke Billing

Detail lengkap di `docs/rules/akuntansi.md` §2 dan `docs/bugs/akuntansi-auto-posting-tidak-terpasang.md`.

1. **Skema**: `payments.debit_account_id` → `NOT NULL` (lihat `docs/schemas/keuangan-akuntansi.md` §3 baris #1). `CreateManualPaymentRequest.DebitAccountID` jadi wajib.
2. **`AutoPostingService`**: perbaiki pencarian periode (pakai rentang tanggal, bukan `FindActive()`), perbaiki kode error mapping akun yang tidak ditemukan, pakai `NextJournalNumber` dari Fase 0.2, tambah pengecekan idempotensi lewat `JournalRepository.FindBySource` sebelum posting.
3. **Wiring**:
   - `CreateInvoiceUseCase`/`CreateInvoiceBatchUseCase` → panggil `PostInvoiceIssued` di transaksi yang sama.
   - `VerifyPaymentUseCase` → panggil `PostPaymentVerified` di transaksi yang sama (sudah pakai `transactor.WithTx`, tinggal disisipkan).
   - `CancelInvoiceUseCase` → panggil `PostInvoiceCancelled` (hanya kalau invoice sebelumnya sudah pernah issued), tambah `transactor.WithTx` (belum ada di use-case ini sekarang).
   - `ApplyAdjustmentUseCase` → posting jurnal koreksi kecil (lihat `docs/rules/akuntansi.md` §2.1), hanya kalau invoice sudah `issued`.
4. **Skema tambahan**: unique partial index `(source_type, source_id)` di `journal_entries` (lihat schema doc §3 baris #3) sebagai jaring pengaman idempotensi di level DB.
5. **API**: lihat `docs/specs/keuangan-akuntansi-api.md` §B.1–B.3, B.6 untuk perubahan perilaku endpoint yang sudah ada, dan §B.4 untuk endpoint baru `GET /admin/journal-entries/by-source`.

**Verifikasi Fase 1**: buat invoice → cek muncul jurnal `invoice_issued` dengan piutang & pendapatan yang benar; verifikasi payment → cek muncul jurnal `payment_verified`; batalkan invoice yang sudah issued → cek muncul jurnal pembalik; ulangi generate massal 2x untuk skema yang sama → tidak boleh dobel-posting jurnal untuk invoice yang sama; buku besar & neraca saldo sekarang menunjukkan saldo yang sesuai transaksi.

---

## FASE 2 — Proses Closing Periode

Detail lengkap di `docs/rules/akuntansi.md` §3.2, `docs/specs/keuangan-akuntansi-api.md` §B.5, `docs/bugs/akuntansi-closing-period-tidak-generate-jurnal.md`.

1. **Skema**: sederhanakan status `accounting_periods` — hapus `closing`, sisakan `open`/`closed`/`locked` (schema doc §3 baris #2). Hapus `AccountingPeriod.StartClosing()` (dead code).
2. **`ClosePeriodUseCase`**: hitung saldo semua akun `revenue`/`expense` di periode itu, buat satu jurnal `source_type='closing'` yang menutup ke `3201 Laba Tahun Berjalan` lalu ke `3200 Saldo Laba`, ubah status periode jadi `closed` — semua dalam satu transaksi.
3. **`ReopenPeriodUseCase`**: batalkan jurnal `closing` periode itu (kalau ada) sebagai bagian dari reopen.
4. Pastikan seed COA (`docs/plan/keuangan-module.md` §"Default Seed Accounts") benar-benar berisi `3200`/`3201` — closing gagal dengan pesan jelas kalau tidak ada, bukan error generik.

**Verifikasi Fase 2**: buat beberapa jurnal (manual + hasil billing dari Fase 1) di satu periode dengan revenue > expense → tutup periode → cek jurnal `closing` muncul dan `3200 Saldo Laba` bertambah sesuai laba; reopen → cek jurnal `closing` ter-cancel dan status kembali `open`.

---

## Catatan untuk Fase Selanjutnya (di luar rencana ini)

- **`finance_audit_logs`** — masih dead table. Keputusan pakai atau tidak ditunda terpisah dari paket ini (lihat `docs/rules/akuntansi.md` §4.8).
- **Mapping `fee_component.type → account.code` yang bisa dikonfigurasi dari UI** (bukan hardcode di `auto_posting.go`) — sudah direncanakan sebagai Fase 2 di `docs/plan/keuangan-module.md`, tetap di luar lingkup dokumen ini.
- **Payment gateway / virtual account** — tetap Fase 3 sesuai `docs/plan/keuangan-module.md`, tidak disentuh di sini.
