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

Urutan kerja konkret (lihat juga `docs/rules/akuntansi.md` §2 dan §3.3 untuk rasionalnya):

### 1. Perbaiki dulu 3 hal kecil di `auto_posting.go` itu sendiri

- **Pencarian periode salah** — `PostInvoiceIssued` (baris 79), `PostPaymentVerified` (baris 125), `PostInvoiceCancelled` (baris 173) semuanya memanggil `s.periodRepo.FindActive(ctx)`. Ganti ketiganya jadi `s.periodRepo.FindByDate(ctx, entryDate)` — method ini **sudah ada dan sudah benar** di `domain/period/repository/interfaces.go:27` serta diimplementasikan di `postgres_period_repo.go:135` (`WHERE start_date<=$1 AND end_date>=$1`), hanya belum pernah dipanggil dari mana pun. Tidak perlu menulis query baru, tinggal ganti nama method yang dipanggil.
- **Kode error salah saat mapping akun tidak ketemu** — baris 67 & 161 mengembalikan `kernel.New(journalConst.CodeJournalNotBalanced)`. Tambah kode baru di `internal/modules/keuangan/domain/journal/constant/journal_constant.go`:
  ```go
  CodeJournalAccountMappingNotFound kernel.Code = "JOURNAL_ACCOUNT_MAPPING_NOT_FOUND"
  ```
  lalu di kedua titik itu: `return kernel.New(journalConst.CodeJournalAccountMappingNotFound)`. Saat use-case pemanggil (lihat langkah 3) menerima error ini, bungkus dengan `application.WrapConflictErr(err, journalConst.CodeJournalAccountMappingNotFound)` supaya jadi HTTP 409, bukan 500 — lihat `docs/bugs/keuangan-kernel-error-tanpa-wrap-selalu-500.md`, jangan sampai bug baru ini mengulang pola yang sama.
- **Nomor jurnal** — ganti `generateJournalNumber(seq, sourceType)` (baris 44-62, yang menerima `seq` mentah dari pemanggil) supaya nomor didapat dari `journalRepo.NextJournalNumber(ctx)` (tambahan baru, lihat `docs/bugs/akuntansi-manual-journal-nomor-selalu-1.md`) alih-alih parameter `seq int`. Hapus parameter `seq int` dari signature keempat method publik (`PostInvoiceIssued`, `PostPaymentVerified`, `PostInvoiceCancelled`) karena tidak dibutuhkan lagi.

### 2. Tambah pengecekan idempotensi di awal tiap method publik

Sebelum membuat jurnal baru, panggil `s.journalRepo.FindBySource(ctx, string(sourceType), sourceID)` — kalau hasilnya bukan error "not found" (artinya sudah pernah diposting), langsung `return nil` (anggap sukses, tidak posting ulang) alih-alih lanjut membuat jurnal baru. Ini mencegah dobel-posting kalau use-case pemanggil di-retry atau (untuk batch) dijalankan ulang.

### 3. Ubah signature & suntikkan `AutoPostingService` ke 4 use-case

| Use-case | Constructor saat ini | Tambahan parameter |
|---|---|---|
| `CreateInvoiceUseCase` (`create_invoice.go:34`) | `NewCreateInvoiceUseCase(invoiceRepo, feeRepo, assignmentRepo, billingPeriodRepo, kesantrianReader, transactor)` | `+ autoPosting *journalService.AutoPostingService` |
| `CreateInvoiceBatchUseCase` (`create_invoice_batch.go:40`) | `NewCreateInvoiceBatchUseCase(invoiceRepo, feeRepo, schemeRepo, assignmentRepo, billingPeriodRepo, batchRepo, targetRepo, kesantrianReader, transactor)` | `+ autoPosting *journalService.AutoPostingService` |
| `VerifyPaymentUseCase` (`verify_payment.go:21`) | `NewVerifyPaymentUseCase(paymentRepo, invoiceRepo, transactor)` | `+ autoPosting *journalService.AutoPostingService` |
| `CancelInvoiceUseCase` (`cancel_invoice.go:16`) | `NewCancelInvoiceUseCase(invoiceRepo)` | `+ invoiceRepo` sudah ada; tambah `+ autoPosting *journalService.AutoPostingService, + transactor ports.Transactor` (use-case ini **belum punya transactor sama sekali** sekarang — wajib ditambah karena sekarang akan menulis ke 2 tabel: `invoices` + `journal_entries`) |

Titik pemanggilan persis:
- `create_invoice.go`, setelah `if cmd.Issue { inv.Issue() }` sukses dan **sebelum** `uc.invoiceRepo.Save(ctx, inv)` — bungkus `Save` + `PostInvoiceIssued` dalam satu `uc.transactor.WithTx(ctx, ...)` baru (use-case ini juga belum pakai transactor untuk `Save`-nya sekarang, walau punya `transactor` di struct-nya — cek dulu, kemungkinan field itu juga belum kepakai).
- `create_invoice_batch.go`, di dalam `processTarget` (baris 174-259), tepat setelah `inv.Issue()` sukses dan `uc.invoiceRepo.Save(txCtx, inv)` sukses (baris 225) — panggilan `PostInvoiceIssued` masuk ke `txCtx` yang sama, di dalam `uc.transactor.WithTx` yang sudah membungkus fungsi ini (baris 187).
- `verify_payment.go`, di dalam `uc.transactor.WithTx` yang sudah ada (baris ~40-46) — tambahkan panggilan `PostPaymentVerified` setelah `uc.invoiceRepo.Update(txCtx, inv)` sukses, sebelum `return nil`.
- `cancel_invoice.go`, bungkus seluruh `Execute` dengan `transactor.WithTx` baru; panggil `PostInvoiceCancelled` hanya kalau `inv.IssuedAt != nil` (artinya pernah issued, pernah menghasilkan jurnal) — sebelum itu, tolak dulu kalau `inv.PaidAmount > 0` (lihat `docs/bugs/akuntansi-cancel-invoice-partial-payment.md`, harus dikerjakan bersamaan supaya tidak ada celah antara dua perbaikan ini).

### 4. Hapus `_ = autoPostingService` di `module.go:102`

Setelah use-case di atas benar-benar menerima `autoPostingService` sebagai parameter constructor (baris yang perlu diubah argumennya: `module.go:56` `createInvoiceUC`, `:57` `createInvoiceBatchUC`, `:58` `cancelInvoiceUC`, `:61` `verifyPaymentUC` — tambahkan `autoPostingService` di akhir daftar argumen tiap panggilan, dan untuk `cancelInvoiceUC` tambahkan juga `transactor`), baris `_ = autoPostingService` di `module.go:102` otomatis tidak diperlukan lagi — hapus.
