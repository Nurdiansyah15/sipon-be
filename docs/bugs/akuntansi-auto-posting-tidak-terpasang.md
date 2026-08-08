# Bug: Auto-posting jurnal tidak pernah terpasang

**Severity**: Tinggi — fitur akuntansi inti (jurnal otomatis dari invoice/payment) secara fungsional tidak ada, meski kodenya sudah ditulis lengkap.

## Lokasi

- `internal/modules/keuangan/domain/journal/service/auto_posting.go` — `AutoPostingService` dengan method `PostInvoiceIssued`, `PostPaymentVerified`, `PostInvoiceCancelled` lengkap dengan logic double-entry yang benar.
- `internal/modules/keuangan/module.go:49` — `autoPostingService := journalService.NewAutoPostingService(journalRepo, accountRepo, periodRepo)` dibuat.
- `internal/modules/keuangan/module.go:102` — **`_ = autoPostingService`** — dibuang ke blank identifier semata-mata supaya compiler tidak komplain "declared and not used".
- `internal/modules/keuangan/application/command/create_invoice.go`, `create_invoice_batch.go`, `verify_payment.go`, `cancel_invoice.go` — tidak satupun menerima/memanggil `AutoPostingService`.

## Gejala

Membuat invoice, memverifikasi payment, atau membatalkan invoice **tidak pernah menghasilkan baris apa pun di `journal_entries`**. Buku besar, neraca saldo, neraca, dan laporan laba rugi akan selalu kosong/nol walau transaksi billing berjalan normal — modul akuntansi secara efektif tidak terhubung ke modul billing sama sekali, padahal itu adalah tujuan utama modul `keuangan` (lihat `docs/plan/keuangan-module.md` §"Keputusan Arsitektur Kunci" poin 1: "setiap aksi billing yang terverifikasi langsung memicu auto-posting jurnal dalam transaksi yang sama").

## Akar Masalah

Auto-posting service dibangun di fase implementasi awal tapi tidak pernah diintegrasikan ke use-case billing yang relevan. `docs/plan/perubahan-skema-tagihan.md` sudah mencatat ini sebagai catatan ("auto_posting.go saat ini kode mati, tidak pernah dipanggil dari manapun — aman diubah/diwiring kapan saja tanpa risiko regresi").

## Dampak

- Tidak ada laporan keuangan akuntansi (neraca, laba rugi, buku besar) yang bisa dipercaya — semuanya akan menunjukkan saldo nol untuk sisi yang seharusnya terisi dari billing.
- Piutang santri (akun `1103`) tidak pernah tercatat meski invoice sudah diterbitkan dan dianggap `issued` di sisi billing.

## Saran Perbaikan

1. Suntikkan `AutoPostingService` (atau interface yang lebih sempit) ke `CreateInvoiceUseCase`, `CreateInvoiceBatchUseCase`, `VerifyPaymentUseCase`, `CancelInvoiceUseCase`.
2. Panggil method posting yang sesuai **di dalam** `transactor.WithTx` yang sudah dipakai use-case tersebut (jangan bikin transaksi terpisah — satu aksi bisnis = satu transaksi DB, termasuk efek jurnalnya).
3. Sebelum posting, cek idempotensi (`JournalRepository.FindBySource`) supaya retry/generate ulang tidak dobel-posting — lihat `docs/rules/akuntansi.md` §2.
4. `PostInvoiceIssued`/`PostInvoiceCancelled` saat ini mengembalikan `kernel.New(journalConst.CodeJournalNotBalanced)` kalau mapping tipe komponen → akun pendapatan tidak ditemukan (`auto_posting.go:67`, `:161`) — kode error ini salah (bukan soal jurnal tidak balance, tapi mapping akun tidak ada). Perbaiki jadi kode error baru yang lebih jelas saat mengerjakan ini, ikuti pola perbaikan kode error yang sama seperti bug "duplikat tagihan" sebelumnya.
5. `PostInvoiceIssued`/`PostPaymentVerified`/`PostInvoiceCancelled` mengambil periode lewat `periodRepo.FindActive(ctx)` (`auto_posting.go:79`, `:125`, `:173`) — ini mengambil periode manapun yang berstatus `open`, bukan periode yang tanggalnya cocok dengan `entryDate`. Perbaiki jadi cari periode lewat rentang tanggal (`start_date <= entryDate <= end_date`) — lihat `docs/rules/akuntansi.md` §3.3.
