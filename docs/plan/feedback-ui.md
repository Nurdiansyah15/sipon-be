# Plan UI: Module Feedback (Umpan Balik) di `sipon-ui`

## Context

Backend module `feedback` sudah diimplementasikan penuh (lihat `docs/plan/feedback-module.md` — modul `internal/modules/feedback/`). Tugas sekarang membangun UI-nya di `sipon-ui` (Nuxt 4), berdasarkan kontrak API **nyata** dari router.go + dto backend.

> **Status: SUDAH DIIMPLEMENTASI** di `sipon-ui` (types, stores, komponen, halaman public + admin, wiring nav). Dokumen ini dokumentasi hidup.

Fitur:
1. User buat feedback (judul, isi, kategori, attachment), komentar flat (bisa reply/tag komentar lain dalam feedback yang sama), like di feedback & komentar.
2. Admin moderasi: takedown/restore feedback & komentar.

## Temuan penting dari kode backend (untuk UI)

1. **List publik hanya feedback `is_takedown=false`.** Detail takedown hanya tampil untuk pemilik (via `/feedbacks/my` atau `/feedbacks/:id` sebagai owner) atau moderator.
2. **`manage_feedback`** adalah permission default system role usergod/superadmin/admin DAN assignable ke custom role. Muncul otomatis di `GET /api/v1/web/role-permission/permission-keys`.
3. **Kategori**: `saran`, `pengaduan`, `pertanyaan`, `apresiasi`, `lainnya`. Default `lainnya`.
4. **Like toggle**: `POST /feedbacks/:id/like` & `POST /comments/:commentId/like` mengembalikan `{liked, like_count}` — tidak perlu state client yang rumit.
5. **Reply komentar**: `POST /feedbacks/:id/comments` dengan body `{body, reply_to_id?}`. `reply_to_id` harus mengacu komentar dalam feedback yang sama. Response list komentar menyertakan `reply_to_id` + `reply_to_user`.
6. **Attachment**: presign → PUT ke storage → confirm (`{key, original_filename, mime_type, size}`). Maks 5 per feedback (backend tolak saat confirm dengan `ATTACHMENT_LIMIT_EXCEEDED`).
7. **Response list** punya `meta` pagination standar; `is_liked` & `is_owner` sudah dihitung backend per user yang login.
8. **Takedown** feedback/komentar menyimpan `takedown_reason` — UI admin wajib kirim reason (textbox) saat takedown.

## Pola existing yang di-reuse

- API layer `useApi()` / `parseApiError()` / envelope types.
- Store Pinia per domain (`app/stores/*.ts`).
- List/table `UTable` + `UPagination` + `usePermission()` inline.
- Presign/confirm upload: pola `AvatarUploadModal.vue`.
- Modal konfirmasi destructive: `ConfirmActionModal.vue`.

## Struktur baru

### Types — `shared/types/Feedback.ts`

```typescript
export type FeedbackCategory = 'saran' | 'pengaduan' | 'pertanyaan' | 'apresiasi' | 'lainnya'

export interface FeedbackUser {
  user_id: string
  username: string
  email: string
  fullname: string | null
}

export interface FeedbackItem {
  id: string
  user?: FeedbackUser
  title: string
  body: string
  category: FeedbackCategory
  is_takedown: boolean
  takedown_reason: string | null
  like_count: number
  comment_count: number
  is_liked: boolean
  attachment_count: number
  created_at: string
  updated_at: string
}

export interface FeedbackAttachment {
  id: string
  key: string
  original_filename?: string
  mime_type?: string
  size?: number
  download_url?: string
  created_at: string
}

export interface FeedbackComment {
  id: string
  feedback_id: string
  user?: FeedbackUser
  body: string
  reply_to_id?: string
  reply_to_user?: FeedbackUser
  is_takedown: boolean
  takedown_reason?: string
  like_count: number
  is_liked: boolean
  is_owner: boolean
  created_at: string
  updated_at: string
}

export interface CreateFeedbackRequest { title: string; body: string; category: FeedbackCategory }
export interface CreateCommentRequest { body: string; reply_to_id?: string }
```

### Stores
- `app/stores/feedback.ts` — self-service: `items`, `meta`, `selected`, `comments`, `attachments`. Actions: `fetchFeedbacks(query)`, `fetchMyFeedbacks()`, `fetchDetail(id)`, `createFeedback(payload)`, `updateFeedback(id, payload)`, `deleteFeedback(id)`, `fetchComments(id)`, `createComment(id, payload)`, `toggleLikeFeedback(id)`, `toggleLikeComment(id)`, `presign/confirm/deleteAttachment`.
- `app/stores/feedbackAdmin.ts` — `manage_feedback`: list semua (termasuk takedown), `takedownFeedback(id, reason)`, `restoreFeedback(id)`, `takedownComment(id, reason)`, `restoreComment(id)`.

### Routing
**Public/self-service (layout `default`, semua user login):**
- `app/pages/feedback/index.vue` — list publik + filter kategori + search
- `app/pages/feedback/create.vue` — form buat feedback + attachment uploader
- `app/pages/feedback/[id]/index.vue` — detail + komentar + form komentar (mode reply) + like
- `app/pages/feedback/my/index.vue` — feedback milik sendiri

**Admin (layout `system-admin`, gated `can('manage_feedback')`):**
- `app/pages/system-admin/feedback/index.vue` — list semua feedback
- `app/pages/system-admin/feedback/[id]/index.vue` — detail + komentar, tombol takedown/restore

### Komponen
- `app/components/feedback/FeedbackCard.vue`, `CategoryBadge.vue`, `LikeButton.vue`, `CommentList.vue`, `CommentForm.vue`, `AttachmentUploader.vue`, `TakedownBanner.vue`.
- `app/components/system-admin/feedback/TakedownModal.vue`, `RestoreModal.vue`.

## Fase Pengerjaan

1. Types (`shared/types/Feedback.ts`) + store self-service.
2. Komponen display bersama (Card, Badge, LikeButton, CommentList/Form).
3. Halaman public + nav "Umpan Balik".
4. Store admin + komponen takedown/restore.
5. Halaman admin.
6. Attachment uploader + integrasi.
7. Polish: empty state, error, responsive.

## Verifikasi

1. Alur user: buat feedback → komentar → reply → like → unlike → edit → delete.
2. Alur admin: list → takedown feedback (reason) → hilang dari publik → restore → muncul lagi; sama untuk komentar.
3. Permission gating: user tanpa `manage_feedback` tidak lihat menu admin.
4. Attachment: upload ≤5 (ke-6 ditolak), hapus, download via `download_url`.
