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

Lihat rancangan lengkap & simplifikasi di `docs/rules/akuntansi.md` §3.2 dan endpoint spec di `docs/specs/keuangan-akuntansi-api.md` §B.5. Ringkasnya:

1. Hapus status `closing` yang tidak pernah dipakai (domain method `StartClosing()` + CHECK constraint DB) — sederhanakan jadi `open → closed → locked`.
2. `ClosePeriodUseCase` membuat satu jurnal `source_type='closing'` yang memindahkan saldo semua akun `revenue`/`expense` periode itu ke `3201 Laba Tahun Berjalan`, lalu ke `3200 Saldo Laba`, dalam transaksi yang sama dengan perubahan status periode.
3. `ReopenPeriodUseCase` membatalkan (bukan menghapus) jurnal `closing` itu kalau ada, sebagai bagian dari reopen.
