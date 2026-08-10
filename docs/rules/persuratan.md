# Rules: Persuratan

Dokumen ini merangkum aturan bisnis modul persuratan (sub-modul kesantrian).

---

## 1. Nomor Surat

Format: `{SEQ}/{KODE_TIPE}/{KODE_ORG}/{ROMAWI}/{YYYY}`

Contoh: `003/SKA/ORG/VIII/2026`

| Komponen | Sumber | Keterangan |
|---|---|---|
| SEQ | Auto-generate | Zero-pad 3 digit, max seq per bulan |
| KODE_TIPE | `master_tipe_surat.kode` | e.g. SKA, SKM |
| KODE_ORG | Constant hardcoded | `OrgCode` di `surat/constant` |
| ROMAWI | Bulan dari `tanggal` | I-XII |
| YYYY | Tahun dari `tanggal` | 4 digit |

### Reset Sequence
- Reset **tiap bulan**, berlaku untuk semua tahun.
- Januari 2026 dan Januari 2027 keduanya mulai dari `001`.
- `SELECT COALESCE(MAX(seq), 0) FROM surat WHERE EXTRACT(MONTH FROM tanggal) = $month AND EXTRACT(YEAR FROM tanggal) = $year`

### Concurrency
- Advisory lock: `pg_advisory_xact_lock(hashtext('persuratan_seq_' || month || '_' || year))`
- Mencegah dua admin mendapat nomor yang sama saat create bersamaan.

### Delete Tidak Reclaim
- Delete surat **tidak** mengurangi seq atau mengosongkan nomor.
- Nomor yang sudah keluar tidak dipakai ulang.

---

## 2. Master Tipe Surat

### Create
- `kode` harus unique, case-sensitive.
- `nama` dan `kode` wajib diisi.

### Update
- Jika tipe sudah dipakai oleh setidaknya satu surat (`surat.tipe_surat_id`):
  - `kode` **tidak boleh diubah** (akan merusak konsistensi nomor).
  - `nama` boleh diubah.
- Jika belum dipakai: `kode` dan `nama` boleh diubah.

### Delete
- Tolak jika sudah dipakai oleh surat manapun.
- Error: 409 Conflict.

---

## 3. Surat

### Create
- `tipe_surat_id` wajib, harus ada di `master_tipe_surat`.
- `tanggal` wajib, format ISO date.
- `keterangan` opsional.
- `dokumen_aset_ids` opsional (array of UUID, bisa kosong).
- Nomor di-generate otomatis di use case.
- `created_by` diambil dari `user_id` (auth context).

### Delete
- Cascade delete ke `surat_dokumen_aset` (FK ON DELETE CASCADE).
- Tidak mempengaruhi `master_tipe_surat` maupun `dokumen_aset`.

### Dokumen Aset Link
- 1 surat bisa punya N dokumen_aset.
- Link via tabel junction `surat_dokumen_aset`.
- `dokumen_aset_id` tidak divalidasi exist di dokumen_aset (cross-module, no FK).
- Unique constraint: `(surat_id, dokumen_aset_id)` — tidak boleh link duplikat.
- Tambah link: POST `/surat/:id/dokumen`.
- Hapus link: DELETE `/surat/:id/dokumen/:dokAsetId`.

---

## 4. Download Dokumen

- Endpoint `/surat/:id/dokumen/:dokAsetId/download` proxy ke `dokumen_aset.Contract.GetDownloadURL`.
- Butuh auth (user terautentikasi) karena dokumen_aset bisa private.
- Response: presigned URL + TTL.

---

## 5. Access Control

- Semua endpoint butuh JWT valid + permission `manage_persuratan`.
- Tidak ada endpoint self-service (user/member).
