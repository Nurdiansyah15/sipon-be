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

Pakai mekanisme yang sama seperti `NextInvoiceNumber`/`NextPaymentNumber` (`internal/modules/keuangan/infrastructure/persistence/number_sequence.go`, tabel `finance_number_sequences`): tambah `NextJournalNumber(ctx) (string, error)` di `JournalRepository`, implementasikan dengan `nextNumberSeq(ctx, execer, "journal", year)`, panggil dari `CreateManualJournalUseCase.Execute` (dan dari `AutoPostingService` sekaligus, ganti `generateJournalNumber(seq, ...)` yang saat ini menerima `seq` mentah dari pemanggil — lihat bug auto-posting).
