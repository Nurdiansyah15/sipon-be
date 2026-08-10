# Plan: Module Feedback (Umpan Balik)

## Context

Sistem SIPON saat ini sudah memiliki modul `identity`, `kesantrian`, `psb`, `dokumen_aset`, `article`, dan `keuangan`. Belum ada modul untuk **umpan balik (feedback)** dari user — saran, pengaduan, pertanyaan, apresiasi.

Tujuan: menambah modul `feedback` dengan fitur:
1. **Feedback** — user buat feedback dengan judul, deskripsi, kategori, attachment
2. **Komentar** — flat (satu level), bisa reply/tag komentar lain dalam feedback yang sama
3. **Like** — di feedback maupun komentar
4. **Moderasi admin** — takedown/restore feedback atau komentar tidak pantas

> **Status: SUDAH DIIMPLEMENTASI** (commit plan ini, modul `internal/modules/feedback/`). Dokumen ini dokumentasi hidup.

## Keputusan Arsitektur

1. **Satu modul `feedback`** — domain kecil, terikat erat.
2. **Cross-module via Contract** — `feedback` memanggil `identity.Contract` untuk resolve nama user (port `ports.IdentityReader` + `infrastructure/identitygateway`). `feedback` belum mengekspos Contract (YAGNI).
3. **Tidak ada FK ke tabel module lain** — `user_id` plain UUID (pola existing).
4. **Flat comments** — satu level; reply/tag lewat `reply_to_id` yang wajib mengacu ke komentar masih-ada & tidak ditakedown dalam feedback yang sama.
5. **Like polymorphic** — satu tabel `feedback_likes` dengan `target_type` (`feedback`|`comment`) + `target_id`.
6. **Moderasi sederhana** — flag `is_takedown` + `takedown_reason` + `takedown_by` + `takedown_at`. Tanpa lifecycle status open/closed.
7. **Identitas selalu tampil** — tidak ada opsi anonim.
8. **Denormalized counters** — `like_count` & `comment_count` di `feedbacks`; `like_count` di `feedback_comments` (update lewat use case, bukan DB trigger).

## Aktor

| Aktor | Peran |
|---|---|
| **User (member)** | Buat feedback, komentar, like, lihat publik, lihat "feedback milik saya" (termasuk yang ditakedown) |
| **Admin / Moderator** | `manage_feedback` — takedown/restore feedback & komentar. Ditambahkan sebagai **default permission** system role `usergod`/`superadmin`/`admin` DAN assignable ke custom role (mis. "moderator"). |

## Skema Data (migration `20260810120000_create_feedback_tables`)

### `feedbacks`
`id UUID PK`, `user_id UUID` (no FK), `title VARCHAR(200)`, `body TEXT`, `category VARCHAR(30)` CHECK (`saran`|`pengaduan`|`pertanyaan`|`apresiasi`|`lainnya`), `is_takedown BOOL`, `takedown_reason TEXT`, `takedown_by UUID`, `takedown_at TIMESTAMPTZ`, `like_count INT`, `comment_count INT`, `created_at/updated_at/deleted_at`.
Index partial publik: `WHERE deleted_at IS NULL AND is_takedown = false`.

### `feedback_attachments`
`id UUID PK`, `feedback_id UUID FK CASCADE`, `key TEXT`, `original_filename VARCHAR(500)`, `mime_type VARCHAR(200)`, `size BIGINT`, `sort_order INT`, timestamps. Maks 5 per feedback (constant `MaxAttachmentsPerFeedback`). Content-type dibatasi (`AllowedContentTypes`).

### `feedback_comments`
`id UUID PK`, `feedback_id UUID FK CASCADE`, `user_id UUID` (no FK), `body TEXT`, `reply_to_id UUID` (self FK, `CHECK (reply_to_id IS NULL OR reply_to_id != id)`), `is_takedown BOOL`, `takedown_reason/by/at`, `like_count INT`, timestamps.

### `feedback_likes`
`id UUID PK`, `user_id UUID`, `target_type VARCHAR(10)` CHECK (`feedback`|`comment`), `target_id UUID`, `created_at`, `UNIQUE (user_id, target_type, target_id)`.

## Business Rules

1. User login (tanpa permission khusus) bisa buat feedback.
2. Edit/delete feedback/komentar hanya oleh pemilik (403 kalau bukan).
3. Reply komentar wajib mengacu ke komentar yang masih ada & tidak ditakedown dalam feedback yang sama (400 kalau beda feedback/soft-deleted/takedown).
4. Like toggle — satu user sekali per target (unique constraint). Unlike men-decrement counter (tidak negatif).
5. Takedown feedback → hilang dari list/detail publik; pemilik tetap bisa lihat sendiri (`/feedbacks/my` & `/feedbacks/:id`); moderator bisa lihat semua.
6. Takedown komentar → hilang dari list komentar publik; tampil untuk admin.
7. Restore membalikkan takedown.
8. Attachment max 5 per feedback; presign (staging `pending/feedback/<id>/...`) → confirm+promote → simpan baris. Hanya pemilik feedback yang bisa upload/hapus attachment.

## Permission

`PermissionManageFeedback = "manage_feedback"` di `permission_constant.go`:
- Masuk `AllPermissionDefinitions` & `DefaultPermissionsInit`.
- Masuk `RolePermissions` usergod, superadmin, admin (default).
- Tetap assignable ke custom role via role management (pola sama `manage_psb`).

## API Endpoints

**Self-service (JWT, tanpa permission khusus):**
- `GET /api/v1/web/feedbacks` — list publik (filter `category`, `search`, pagination)
- `GET /api/v1/web/feedbacks/my` — milik sendiri (termasuk ditakedown)
- `GET /api/v1/web/feedbacks/:id` — detail + attachments
- `POST /api/v1/web/feedbacks` — buat
- `PUT /api/v1/web/feedbacks/:id` — edit (pemilik)
- `DELETE /api/v1/web/feedbacks/:id` — hapus (pemilik)
- `GET/POST /api/v1/web/feedbacks/:id/comments` — list & buat komentar
- `PUT/DELETE /api/v1/web/comments/:commentId` — edit/hapus komentar (pemilik)
- `POST /api/v1/web/feedbacks/:id/like` — toggle like feedback
- `POST /api/v1/web/comments/:commentId/like` — toggle like komentar
- `POST /api/v1/web/feedbacks/:id/attachments/presign|confirm`, `GET /api/v1/web/feedbacks/:id/attachments`, `DELETE /api/v1/web/feedbacks/:id/attachments/:attachmentId`

**Admin (`manage_feedback`):**
- `GET /api/v1/web/feedback/admin/feedbacks` (termasuk ditakedown)
- `GET /api/v1/web/feedback/admin/feedbacks/:id`
- `GET /api/v1/web/feedback/admin/feedbacks/:id/comments` (termasuk ditakedown)
- `POST /api/v1/web/feedback/admin/feedbacks/:id/takedown|restore`
- `POST /api/v1/web/feedback/admin/comments/:commentId/takedown|restore`

## Struktur Module

```
internal/modules/feedback/
├── module.go
├── domain/
│   ├── feedback/{constant,entity,repository}
│   ├── comment/{constant,entity,repository}
│   ├── like/{constant,entity,repository}
│   └── attachment/{constant,entity,repository}
├── application/
│   ├── command/  -- create/update/delete/moderate feedback, comments, toggle_like, attachment
│   ├── query/    -- list_feedbacks, get_feedback, list_comments, list_attachments (+helpers)
│   ├── dto/
│   └── ports/    -- identity_reader, storage, transactor
├── infrastructure/
│   ├── persistence/  -- postgres repos + transactor + helpers
│   ├── external/     -- minio_uploader
│   └── identitygateway/
└── interfaces/http/
    ├── handler.go
    └── router.go
```

## Verifikasi

1. `go build ./...`, `go vet ./...`, `go test ./internal/modules/feedback/...` lolos.
2. Migration `up`/`down` bersih (sudah diuji di dev).
3. Smoke test E2E (sudah dijalankan): create feedback → komentar → reply dalam feedback sama (OK) / reply lintas feedback (400) → like feedback & komentar → unlike → edit/delete pemilik → non-owner edit/delete (403) → takedown feedback (hilang dari publik, muncul di admin) → takedown komentar → restore → muncul lagi.
