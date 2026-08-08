# Bug: Laporan ledger & trial balance query kolom `deleted_at` yang tidak ada

**Severity**: Tinggi — dua laporan akuntansi akan selalu gagal (error DB), bukan sekadar menampilkan data kosong.

## Lokasi

- `internal/modules/keuangan/application/query/report_ledger.go` — query SQL berisi `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL`.
- `internal/modules/keuangan/application/query/report_trial_balance.go` — query SQL berisi `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL`.
- Skema aktual (`migrations/*_create_keuangan_tables.up.sql`, tabel `journal_entries` & `journal_entry_lines`): **tidak ada kolom `deleted_at`** di kedua tabel ini. Jurnal tidak pernah soft-delete — pembatalan dilakukan lewat `status='cancelled'` (lihat `JournalEntry.Cancel()`), bukan penghapusan.

## Gejala

Memanggil `GET /admin/reports/ledger` atau `GET /admin/reports/trial-balance` akan menghasilkan error dari Postgres (`column "deleted_at" does not exist`), diteruskan sebagai 500 ke klien. Kedua laporan ini **tidak bisa dipakai sama sekali** dalam kondisi sekarang, terlepas dari isu auto-posting yang belum terpasang.

## Akar Masalah

Kemungkinan disalin dari pola query tabel lain di modul ini yang memang punya soft-delete (mis. `invoices`, `accounts`), tanpa menyesuaikan ke skema `journal_entries`/`journal_entry_lines` yang tidak memilikinya.

## Dampak

Dua dari enam laporan akuntansi (buku besar & neraca saldo) tidak berfungsi.

## Saran Perbaikan

Hapus klausa `AND jel.deleted_at IS NULL AND je.deleted_at IS NULL` dari kedua query. Sebagai gantinya, filter yang benar untuk "jurnal yang valid dihitung" adalah **`je.status = 'posted'`** (lihat `docs/rules/akuntansi.md` §4.7) — tambahkan itu kalau belum ada, supaya jurnal `draft` (seharusnya tidak pernah ada karena `Post()` dipanggil sebelum `Save()`, tapi tetap sebagai jaga-jaga) dan `cancelled` tidak ikut terhitung di laporan.
