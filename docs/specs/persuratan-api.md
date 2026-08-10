# API Spec: Persuratan

Base path: `/api/v1/web/santri/admin/persuratan`. Semua endpoint butuh JWT + permission `manage_persuratan`.

---

## Tipe Surat

### List Tipe Surat
```
GET /tipe-surat
```
Response: array of `TipeSuratResponse`

### Detail Tipe Surat
```
GET /tipe-surat/:id
```
Response: `TipeSuratResponse` atau 404

### Buat Tipe Surat
```
POST /tipe-surat
{
  "nama": "Surat Keterangan Aktif",
  "kode": "SKA"
}
```
Response: 201 + `TipeSuratResponse`
Error: 409 jika `kode` sudah ada

### Update Tipe Surat
```
PUT /tipe-surat/:id
{
  "nama": "Surat Keterangan Baru",
  "kode": "SKB"
}
```
Response: 200 + `TipeSuratResponse`
Error: 409 jika sudah dipakai surat, 404 jika tidak ada

### Hapus Tipe Surat
```
DELETE /tipe-surat/:id
```
Response: 200
Error: 409 jika sudah dipakai surat, 404 jika tidak ada

---

## Surat

### List Surat
```
GET /surat?page=1&limit=10&tipe_surat_id=&bulan=&tahun=&search=
```
Response: paginasi `SuratResponse` + meta

### Detail Surat
```
GET /surat/:id
```
Response: `SuratDetailResponse` (include `dokumen_aset_ids[]`)
Error: 404 jika tidak ada

### Buat Surat
```
POST /surat
{
  "tipe_surat_id": "uuid",
  "keterangan": "Surat keterangan aktif atas nama ...",
  "tanggal": "2026-08-10",
  "dokumen_aset_ids": ["uuid1", "uuid2"]
}
```
Response: 201 + `SuratDetailResponse` (nomor auto-generated)
Error: 404 jika tipe_surat_id tidak ada

### Hapus Surat
```
DELETE /surat/:id
```
Response: 200
Error: 404 jika tidak ada

### Tambah Dokumen ke Surat
```
POST /surat/:id/dokumen
{
  "dokumen_aset_id": "uuid"
}
```
Response: 200 + `TautDokumenResponse`
Error: 404 jika surat tidak ada, 409 jika sudah terlink

### Hapus Dokumen dari Surat
```
DELETE /surat/:id/dokumen/:dokumenAsetId
```
Response: 200
Error: 404 jika link tidak ada

### Download Dokumen dari Surat
```
GET /surat/:id/dokumen/:dokumenAsetId/download
```
Response: `{ "access_url": "...", "expires_in": 300 }`
Error: 404 jika surat/dokumen tidak ada

---

## Response Shapes

### TipeSuratResponse
```json
{
  "id": "uuid",
  "nama": "Surat Keterangan Aktif",
  "kode": "SKA",
  "created_by": "uuid",
  "created_at": "2026-08-10T12:00:00Z",
  "updated_at": "2026-08-10T12:00:00Z"
}
```

### SuratResponse
```json
{
  "id": "uuid",
  "nomor": "003/SKA/ORG/VIII/2026",
  "tipe_surat_id": "uuid",
  "keterangan": "...",
  "tanggal": "2026-08-10",
  "created_by": "uuid",
  "created_at": "2026-08-10T12:00:00Z"
}
```

### SuratDetailResponse
```json
{
  "id": "uuid",
  "nomor": "003/SKA/ORG/VIII/2026",
  "tipe_surat_id": "uuid",
  "tipe_surat_nama": "Surat Keterangan Aktif",
  "tipe_surat_kode": "SKA",
  "keterangan": "...",
  "tanggal": "2026-08-10",
  "created_by": "uuid",
  "created_at": "2026-08-10T12:00:00Z",
  "dokumen_aset_ids": ["uuid1", "uuid2"]
}
```
