# Bugs

Daftar bug konkret yang ditemukan di modul (bukan rencana fitur — untuk itu lihat `docs/plan/`). Satu file per bug, judul file singkat & deskriptif. Tambahkan entri baru di sini saat menemukan bug baru; jangan duplikasi entri yang sudah ada — cek dulu apakah sudah tertulis.

Format tiap file: **Lokasi** (file:line), **Gejala**, **Akar masalah**, **Dampak**, **Saran perbaikan**.

## Akuntansi / Keuangan

| Bug | Severity | Status |
|---|---|---|
| [Auto-posting jurnal tidak pernah terpasang](akuntansi-auto-posting-tidak-terpasang.md) | Tinggi | Diperbaiki |
| [Nomor jurnal manual selalu 1](akuntansi-manual-journal-nomor-selalu-1.md) | Tinggi | Diperbaiki |
| [Baris jurnal (journal_entry_lines) tidak pernah tersimpan](akuntansi-journal-lines-tidak-tersimpan.md) | Tinggi | Diperbaiki |
| [4 dari 6 laporan akuntansi query kolom `deleted_at` yang tidak ada](akuntansi-laporan-jurnal-kolom-deleted-at-tidak-ada.md) | Tinggi | Diperbaiki |
| [Error domain tanpa Wrap selalu jadi HTTP 500 (9 lokasi)](keuangan-kernel-error-tanpa-wrap-selalu-500.md) | Tinggi | Diperbaiki |
| [Proses closing periode tidak membuat jurnal penutup](akuntansi-closing-period-tidak-generate-jurnal.md) | Sedang | Diperbaiki |
| [Cancel invoice tidak menolak invoice berstatus partial](akuntansi-cancel-invoice-partial-payment.md) | Sedang | Diperbaiki |
| [Cancel journal tidak mengecek status periode akuntansi](akuntansi-cancel-journal-tidak-cek-periode.md) | Sedang | Diperbaiki |
