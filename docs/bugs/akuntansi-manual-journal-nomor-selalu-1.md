# Bug: Nomor jurnal manual selalu di-hardcode seq=1

**Severity**: Tinggi — endpoint jurnal manual yang sudah aktif (`POST /admin/journal-entries`) akan gagal dipakai lebih dari satu kali dalam bulan yang sama.

## Lokasi

`internal/modules/keuangan/application/command/create_manual_journal.go:49`

```go
jn := fmt.Sprintf("JRN/%d/%02d/%06d", time.Now().Year(), time.Now().Month(), 1)
```

## Gejala

Nomor jurnal manual pertama di bulan tertentu akan sukses (`JRN/2026/08/000001`). Jurnal manual **kedua** di bulan yang sama akan mencoba insert dengan nomor yang identik (`seq` selalu `1`) — gagal karena `journal_number` `UNIQUE` di tabel `journal_entries`, kemungkinan besar muncul sebagai error 500 generik ke bendahara.

Ini persis pola bug yang sama seperti `invoice_number`/`payment_number` sebelum diperbaiki (lihat `docs/plan/perubahan-skema-tagihan.md` poin bug #2, dan perbaikannya lewat tabel `finance_number_sequences`) — hanya saja untuk jurnal manual perbaikannya belum pernah dilakukan.

## Akar Masalah

Nomor dihitung langsung di application layer dengan angka tetap `1`, bukan lewat sequence atomic di database.

## Dampak

Fitur jurnal manual (satu-satunya cara bendahara mencatat transaksi non-billing seperti gaji, biaya operasional) pada praktiknya hanya bisa dipakai sekali per bulan sebelum mulai gagal.

## Saran Perbaikan

Pakai mekanisme yang sama seperti `NextInvoiceNumber`/`NextPaymentNumber` (`internal/modules/keuangan/infrastructure/persistence/number_sequence.go`, tabel `finance_number_sequences`):

1. **Tambah ke interface** `internal/modules/keuangan/domain/journal/repository/interfaces.go`, di dalam `JournalRepository`:
   ```go
   NextJournalNumber(ctx context.Context) (string, error)
   ```
2. **Implementasikan** di `internal/modules/keuangan/infrastructure/persistence/postgres_journal_repo.go` (pola persis `PostgresInvoiceRepository.NextInvoiceNumber`, `postgres_invoice_repo.go:31-39`):
   ```go
   func (r *PostgresJournalRepository) NextJournalNumber(ctx context.Context) (string, error) {
       execer := execerFromContext(ctx, r.db)
       year := time.Now().Year()
       seq, err := nextNumberSeq(ctx, execer, "journal", year)
       if err != nil {
           return "", kernel.Wrap(constant.CodeJournalPersistenceFailed, err)
       }
       return fmt.Sprintf("JRN/%d/%02d/%06d", year, int(time.Now().Month()), seq), nil
   }
   ```
   (`nextNumberSeq` sudah ada dan reusable lintas repo — sama fungsi yang dipakai `NextInvoiceNumber`/`NextPaymentNumber` — cukup panggil dengan `doc_type="journal"`.)
3. **Ganti pemanggil**:
   - `create_manual_journal.go:49` — ganti `jn := fmt.Sprintf("JRN/%d/%02d/%06d", ...)` jadi `jn, err := uc.journalRepo.NextJournalNumber(ctx)` (tangani error dengan `application.WrapRepoErr(err, journalConst.CodeJournalPersistenceFailed)`, ikuti pola penanganan error `NextInvoiceNumber` di `create_invoice.go:89-92`).
   - `auto_posting.go` — ganti `generateJournalNumber(seq, sourceType)` (baris 44-62) jadi panggilan `s.journalRepo.NextJournalNumber(ctx)`, lalu tempelkan prefix kosmetik sesuai `sourceType` secara terpisah kalau prefix per jenis sumber (`INV`/`PAY`/`CNL`/`ADJ`/`CLS`/`JRN`) masih mau dipertahankan — lihat `docs/schemas/keuangan-akuntansi.md` §3 baris #5 untuk opsi menyederhanakannya jadi satu prefix `JRN` untuk semua (lebih ringkas, tidak kehilangan informasi karena `source_type` sudah tersimpan terpisah di kolom `journal_entries.source_type`).

