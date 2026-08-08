# Bug: Proses closing periode tidak membuat jurnal penutup (dan status `closing` dead code)

**Severity**: Sedang — belum ada jalur yang mengklaim fitur ini "sudah jadi" secara eksplisit di kode aktif, tapi ini gap fungsional besar terhadap rencana awal yang sudah didokumentasikan.

## Lokasi

- `internal/modules/keuangan/application/command/close_period.go` — `Execute` hanya memanggil `period.Close(closedBy)` lalu `Update`, tidak membuat jurnal apa pun.
- `internal/modules/keuangan/domain/period/entity/period.go:78-85` — method `StartClosing()` (transisi `open → closing`) ada di domain, tapi **tidak ada use-case manapun yang memanggilnya** — dead code.
- `docs/plan/keuangan-module.md` §"Auto-Posting Rules" → "Rule 4: Closing Periode" mendokumentasikan rencana lengkap jurnal penutup (pindah saldo revenue/expense ke Laba Tahun Berjalan → Saldo Laba), tapi tidak ada implementasinya di `application/command/` maupun `domain/journal/service/`.

## Gejala

Menutup periode akuntansi (`POST /admin/periods/:id/close`) hanya mengubah `status` jadi `closed` dan mengisi `closed_by`/`closed_at`. Saldo akun Pendapatan & Beban **tidak pernah dipindahkan** ke ekuitas — kalau periode berikutnya dibuka, laporan laba rugi periode baru akan tercampur dengan saldo revenue/expense periode lama (karena tidak pernah "ditutup ke nol").

## Akar Masalah

Rule 4 dari rencana awal (`docs/plan/keuangan-module.md`) tidak pernah diimplementasikan — kemungkinan ditunda karena bergantung pada auto-posting (yang sendirinya belum terpasang, lihat bug terkait) sudah berjalan lebih dulu.

## Dampak

Laporan laba rugi & neraca per periode tidak akurat setelah lebih dari satu periode berjalan, karena saldo nominal (revenue/expense) tidak pernah direset ke ekuitas.

## Saran Perbaikan

Algoritma lengkap & bukti keseimbangan debit/kredit ada di `docs/rules/akuntansi.md` §3.2 — di sini fokus ke perubahan kode konkret:

1. **Hapus status `closing`**: hapus `AccountingPeriod.StartClosing()` (`period.go:78-85`), hapus nilai `closing` dari CHECK constraint `accounting_periods.status` (migrasi baru, lihat `docs/schemas/keuangan-akuntansi.md` §"accounting_periods"). Sisakan `open → closed → locked`.
2. **Ubah constructor `ClosePeriodUseCase`** (`close_period.go:12-18`, saat ini `NewClosePeriodUseCase(periodRepo)`) jadi:
   ```go
   func NewClosePeriodUseCase(
       periodRepo periodRepo.AccountingPeriodRepository,
       accountRepo accRepo.AccountRepository,
       journalRepo journalRepo.JournalRepository,
       transactor ports.Transactor,
   ) *ClosePeriodUseCase
   ```
   `Execute` dibungkus `transactor.WithTx`, melakukan langkah 1-4 di `docs/rules/akuntansi.md` §3.2 (validasi akun 3200/3201 ada → hitung saldo revenue/expense via query yang sama dengan `report_income_statement.go` → susun & `Save`+`SaveLines` jurnal `closing` kalau ada aktivitas → `period.Close(closedBy)` + `periodRepo.Update`).
3. **Ubah constructor `ReopenPeriodUseCase`** (`reopen_period.go:12-18`, saat ini `NewReopenPeriodUseCase(periodRepo)`) jadi menerima juga `journalRepo` + `transactor`. `Execute` dibungkus `transactor.WithTx`: `journalRepo.FindBySource(ctx, "closing", periodID)` — kalau ketemu (bukan error not-found), `entry.Cancel()` + `journalRepo.Update`; lanjut `period.Reopen()` + `periodRepo.Update` seperti sekarang.
4. **Update wiring di `module.go`** (baris `closePeriodUC := command.NewClosePeriodUseCase(periodRepo)` dan `reopenPeriodUC := command.NewReopenPeriodUseCase(periodRepo)`) — tambahkan argumen baru sesuai signature di atas (`accountRepo`, `journalRepo`, `transactor` sudah semuanya sudah dibuat lebih awal di fungsi yang sama, tinggal pass).
