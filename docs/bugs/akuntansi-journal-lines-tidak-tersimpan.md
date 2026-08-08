# Bug: `journal_entry_lines` tidak pernah tersimpan

**Severity**: Tinggi — jurnal yang dibuat (manual maupun otomatis nanti) tidak punya rincian baris debit/kredit di database, walau header-nya sukses tersimpan.

## Lokasi

- `internal/modules/keuangan/infrastructure/persistence/postgres_journal_repo.go` — method `Save(ctx, entry)` (baris ~30) hanya melakukan `INSERT INTO journal_entries (...)`. Method `SaveLines(ctx, entryID, lines)` (baris ~182) ada dan berfungsi, tapi **tidak pernah dipanggil** dari `Save` maupun dari pemanggil manapun.
- Dikonfirmasi lewat pencarian: `grep -rn "SaveLines(" internal` hanya menemukan definisi interface & implementasi, nol pemanggilan.
- Dipakai oleh: `internal/modules/keuangan/application/command/create_manual_journal.go:80` (`uc.journalRepo.Save(ctx, entry)` — `entry.Lines` yang sudah dikumpulkan lewat `entry.AddLine()` tidak ikut tersimpan) dan `auto_posting.go` (tiga tempat, sama-sama hanya `journalRepo.Save`).

## Gejala

Setelah membuat jurnal manual (`POST /admin/journal-entries`), response API menunjukkan `lines` dengan benar (karena dibangun dari objek in-memory `entry.Lines`, bukan dibaca ulang dari DB) — tapi kalau jurnal itu diambil ulang lewat `GET /admin/journal-entries/:id` (yang memanggil `FindLinesByEntryID`), baris jurnalnya akan **kosong**. Laporan apa pun yang bergantung pada `journal_entry_lines` (buku besar, neraca saldo — lihat bug terkait) tidak akan pernah melihat transaksi ini sama sekali walau header jurnalnya ada dan `total_debit`/`total_credit` di header tercatat benar.

## Akar Masalah

`Save` di `PostgresJournalRepository` hanya menangani tabel header, lupa memanggil `SaveLines` untuk baris-barisnya.

## Dampak

Setiap jurnal yang pernah dibuat lewat `CreateManualJournalUseCase` (satu-satunya jalur jurnal yang sudah aktif saat ini) punya `total_debit`/`total_credit` di header tapi nol baris rincian — data secara akuntansi tidak lengkap/tidak bisa ditelusuri per akun.

## Saran Perbaikan

Di `PostgresJournalRepository.Save`, setelah insert header sukses, panggil `r.SaveLines(ctx, entry.ID, entry.Lines)` sebelum return (pakai `execer`/context transaksi yang sama, supaya tetap atomik dalam `WithTx` yang membungkusnya). Perhatikan bahwa `CreateManualJournalUseCase` saat ini **tidak** dibungkus `transactor.WithTx` sama sekali (tidak ada `ports.Transactor` yang di-inject) — kalau `Save` sukses tapi `SaveLines` gagal di tengah jalan tanpa transaksi, akan ada jurnal header "yatim" tanpa baris. Sebaiknya sekalian tambahkan `transactor.WithTx` ke use-case ini saat memperbaiki bug ini.
