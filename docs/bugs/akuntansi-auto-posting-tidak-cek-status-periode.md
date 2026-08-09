# Bug: `AutoPostingService` tidak menolak posting ke periode akuntansi tertutup

**Severity**: Tinggi — bisa merusak periode yang sudah difinalisasi tanpa ada error apa pun.

> **Status: Diperbaiki** — `AutoPostingService` kini memakai helper `findPostablePeriod` (gabungan `FindByDate` + cek `period.CanPost()`); keempat method menolak posting ke periode tertutup/dikunci dengan `CodeJournalPeriodClosed` ("Periode akuntansi untuk tanggal ini sudah ditutup"). Pemanggil HTTP (`create_invoice`, `verify_payment`, `cancel_invoice`, `apply_adjustment`) memetakan error itu ke 409; batch memunculkannya sebagai reason target.

## Lokasi

`internal/modules/keuangan/domain/journal/service/auto_posting.go` — keempat method (`PostInvoiceIssued`, `PostPaymentVerified`, `PostInvoiceCancelled`, `PostAdjustment`) memanggil `s.periodRepo.FindByDate(ctx, entryDate)` lalu **langsung** lanjut membuat & menyimpan jurnal tanpa mengecek status periode yang ditemukan.

`FindByDate` (`postgres_period_repo.go:135-142`) sendiri cuma `WHERE start_date<=$1 AND end_date>=$1` — tidak peduli `status` (`open`/`closed`/`locked`).

## Gejala

Payment dengan `payment_date` yang jatuh di rentang tanggal periode akuntansi yang sudah `closed`/`locked` (bisa terjadi kalau bendahara mengisi tanggal mundur, atau invoice yang tanggal terbitnya jatuh di periode yang ditutup lebih awal dari akhir tahun kalender) tetap **berhasil** diposting ke jurnal — masuk ke periode yang seharusnya sudah final. Beda dengan jurnal manual (`CreateManualJournalUseCase`) yang sudah benar mengecek `period.CanPost()` sebelum posting.

## Dampak

Laporan periode yang sudah ditutup/dikunci (dan mungkin sudah dilaporkan ke yayasan) bisa berubah lagi diam-diam tanpa jejak peringatan apa pun — bertentangan langsung dengan tujuan closing (`docs/rules/akuntansi.md` §3).

## Saran Perbaikan

Tambah di keempat method, tepat setelah `FindByDate`:
```go
period, err := s.periodRepo.FindByDate(ctx, entryDate)
if err != nil {
    return fmt.Errorf("find active period: %w", err)
}
if !period.CanPost() {
    return kernel.WrapMsg(journalConst.CodeJournalPeriodClosed, "Periode akuntansi untuk tanggal ini sudah ditutup", nil)
}
```
Lalu tambah `case journalConst.CodeJournalPeriodClosed:` di error-switch tiap pemanggil (`create_invoice.go`, `create_invoice_batch.go`, `verify_payment.go`, `cancel_invoice.go`, `apply_adjustment.go`) → `kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)` (409), pola yang sama seperti `CodeJournalAccountMappingNotFound` yang sudah ada di file-file itu.

Dikerjakan sebagai Fase 0 di `docs/plan/invoice-issued-date-dan-periode-opsional.md`.
