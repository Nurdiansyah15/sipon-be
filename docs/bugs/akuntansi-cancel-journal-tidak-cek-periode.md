# Bug: Cancel journal tidak mengecek status periode akuntansi

**Severity**: Sedang.

## Lokasi

`internal/modules/keuangan/application/command/cancel_journal.go`

```go
func (uc *CancelJournalUseCase) Execute(ctx context.Context, journalID string) (*dto.MessageResponse, error) {
	entry, err := uc.journalRepo.FindByID(ctx, journalID)
	...
	if err := entry.Cancel(); err != nil { ... }
	if err := uc.journalRepo.Update(ctx, entry); err != nil { ... }
	...
}
```

Bandingkan dengan `CreateManualJournalUseCase.Execute` (`create_manual_journal.go:40`) yang **sudah benar** mengecek `period.CanPost()` sebelum membuat jurnal baru.

## Gejala

Jurnal manual yang sudah diposting di periode yang sekarang berstatus `closed` atau `locked` masih bisa dibatalkan lewat `POST /admin/journal-entries/:id/cancel` — tidak ada pengecekan status periode sama sekali, hanya status jurnal itu sendiri (`JournalDraft`/`JournalPosted`/`JournalCancelled`) dan `SourceType` (harus `manual`).

## Akar Masalah

`CancelJournalUseCase` tidak pernah diberi akses ke `AccountingPeriodRepository`, jadi tidak ada tempat untuk mengecek status periode dari jurnal yang mau dibatalkan.

## Dampak

Melanggar prinsip "periode yang sudah ditutup/dikunci tidak boleh berubah lagi" (`docs/rules/akuntansi.md` §3.1, §4.5) — laporan periode yang sudah difinalisasi/dikirim ke yayasan bisa berubah diam-diam kalau ada jurnal di periode itu yang dibatalkan belakangan.

## Saran Perbaikan

Suntikkan `periodRepo periodRepo.AccountingPeriodRepository` ke `CancelJournalUseCase`, ambil periode dari `entry.PeriodID`, tolak kalau `!period.IsOpen()` — konsisten dengan pengecekan yang sudah ada di alur pembuatan jurnal.
