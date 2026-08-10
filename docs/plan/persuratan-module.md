# Plan: Modul Persuratan (Kesantrian)

## Context

Sistem SIPON belum memiliki modul untuk mengelola **surat-menyurat** yang dikeluarkan oleh sekretariat. Modul ini menangani:
1. **Master Tipe Surat** — CRUD jenis surat (nama + kode)
2. **Surat** — pencatatan surat keluar dengan nomor otomatis, keterangan, dan link ke dokumen aset (template maupun hasil)

Semua endpoint **admin-only**, dilindungi permission `manage_persuratan`.

> **Status: SUDAH DIIMPLEMENTASI** — sub-modul dalam `internal/modules/kesantrian/`. Dokumen ini dokumentasi hidup.

## Keputusan Arsitektur

1. **Sub-modul di dalam `kesantrian`** — bukan module terpisah, karena persuratan adalah fungsi sekretariat yang terikat erat dengan kesantrian.
2. **Nomor surat auto-generate** — format `{SEQ}/{KODE_TIPE}/{KODE_ORG}/{ROMAWI}/{YYYY}`, seq reset tiap bulan.
3. **Kode organisasi** — hardcoded constant `OrgCode` di `domain/surat/constant`.
4. **Surat bersifat immutable** — hanya create + delete, tidak ada edit/update. Nomor yang sudah keluar tidak boleh berubah.
5. **1 surat : N dokumen aset** — lewat tabel junction `surat_dokumen_aset`. Bisa link template maupun hasil final.
6. **Dokumen_aset dikelola terpisah** — upload/download tetap lewat modul `dokumen_aset`. Persuratan hanya menyimpan referensi `dokumen_aset_id`.
7. **Sekretaris** — direkam otomatis dari `user_id` (auth), tidak ada field nama terpisah.
8. **Cross-module download** — `dokumen_aset` expose `Contract` dengan method `GetDownloadURL`. Kesantrian memanggil lewat parameter constructor.
9. **Concurrency** — nomor surat menggunakan `pg_advisory_xact_lock` per (bulan, tahun) untuk mencegah duplikasi saat concurrent create.

## Skema Data (migration `20260810130000_create_persuratan_tables`)

### `master_tipe_surat`
`id UUID PK`, `nama VARCHAR(200) NOT NULL`, `kode VARCHAR(20) UNIQUE NOT NULL`, `created_by UUID`, `created_at/updated_at TIMESTAMPTZ`.

### `surat`
`id UUID PK`, `nomor VARCHAR(100) UNIQUE NOT NULL`, `seq INT NOT NULL`, `tipe_surat_id UUID FK → master_tipe_surat`, `keterangan TEXT`, `tanggal DATE NOT NULL`, `created_by UUID NOT NULL`, `created_at/updated_at TIMESTAMPTZ`.

### `surat_dokumen_aset`
`id UUID PK`, `surat_id UUID FK → surat ON DELETE CASCADE`, `dokumen_aset_id UUID NOT NULL`, `created_at TIMESTAMPTZ`, `UNIQUE (surat_id, dokumen_aset_id)`.

## Business Rules

1. Admin dengan `manage_persuratan` bisa CRUD tipe surat dan create/delete surat.
2. Tipe surat yang sudah dipakai (ada di `surat`) tidak bisa dihapus atau diubah kodenya.
3. Nomor surat: `{SEQ}/{KODE_TIPE}/{KODE_ORG}/{ROMAWI}/{YYYY}` — seq zero-pad 3 digit, reset tiap bulan.
4. Delete surat tidak reclaim seq — nomor yang sudah keluar tidak dipakai ulang.
5. Dokumen_aset yang di-link tidak harus ada saat create surat (cross-module, tidak ada FK).

## Permission

`PermissionManagePersuratan = "manage_persuratan"` di `permission_constant.go`:
- Masuk `AllPermissionDefinitions` & `DefaultPermissionsInit`
- Masuk `RolePermissions` usergod, superadmin, admin
- Assignable ke custom role

## API Endpoints

**Base: `/api/v1/web/santri/admin/persuratan`** — semua butuh JWT + `manage_persuratan`

### Tipe Surat
| Method | Path | Deskripsi |
|---|---|---|
| GET | `/tipe-surat` | List semua tipe surat |
| GET | `/tipe-surat/:id` | Detail tipe surat |
| POST | `/tipe-surat` | Buat tipe surat baru |
| PUT | `/tipe-surat/:id` | Update tipe surat (tolak jika sudah dipakai) |
| DELETE | `/tipe-surat/:id` | Hapus (tolak jika sudah dipakai) |

### Surat
| Method | Path | Deskripsi |
|---|---|---|
| POST | `/surat` | Buat surat baru (auto-generate nomor) |
| GET | `/surat` | List surat (paginate, filter tipe/bulan/tahun) |
| GET | `/surat/:id` | Detail surat + dokumen_aset terlink |
| DELETE | `/surat/:id` | Hapus surat |
| POST | `/surat/:id/dokumen` | Tambah link dokumen_aset |
| DELETE | `/surat/:id/dokumen/:dokAsetId` | Hapus link dokumen_aset |
| GET | `/surat/:id/dokumen/:dokAsetId/download` | Get download URL (proxy ke dokumen_aset) |

## Struktur Module

```
internal/modules/kesantrian/
├── domain/
│   ├── tipe_surat/{entity,constant,repository}
│   └── surat/{entity,constant,repository,service}
├── application/
│   ├── command/   -- create/update/delete tipe_surat, create/delete surat, add/remove dokumen
│   ├── query/     -- list/get tipe_surat, list/get surat, download proxy
│   └── dto/       -- persuratan DTOs
├── infrastructure/
│   ├── persistence/   -- postgres_tipe_surat_repo, postgres_surat_repo
│   └── dokumenasetgateway/  -- adapter ke dokumen_aset.Contract
├── interfaces/http/
│   ├── persuratan_handler.go
│   └── persuratan_router.go
├── contract.go     -- (existing, unchanged)
└── module.go       -- wire persuratan
```

## Verifikasi

1. `go build ./...`, `go vet ./...` lolos.
2. Migration `up`/`down` bersih.
3. Smoke test E2E: create tipe surat → create surat (cek nomor) → tambah dokumen → download → delete surat → cek seq tidak reclaim.
