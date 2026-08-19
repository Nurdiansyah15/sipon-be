# Plan UI: Module Notification (In-App + Preferensi) di `sipon-ui`

## Context

Backend module `notification` fase 1 sudah diimplementasikan (lihat
`docs/plan/notification-module.md` — modul `internal/modules/notification/`).
Tugas sekarang membangun UI-nya di `sipon-ui` (Nuxt 4), fokus pada **dua** hal:

1. **In-app notification** — bell icon dengan badge jumlah belum-dibaca, panel/dropdown
   list notifikasi, halaman inbox penuh, tandai dibaca (satu & semua).
2. **Preferensi notifikasi** — global on/off (`all_notifications_enabled`) +
   Do-Not-Disturb (`do_not_disturb` + rentang waktu).

> **Tidak termasuk** di sini: admin broadcast (permission `manage_notification`),
> channel email/sms/push (fase 2), realtime (fase 3). Semua di luar cakupan plan ini.

## Kontrak API nyata (dari backend yang sudah dibuat)

Semua endpoint JWT auth (`jwtAuth + principalLoad`), tanpa permission key.
Envelope standar `ApiSuccess<T>` (`{ status, status_code, message, data, meta }`).

| Endpoint | Response `data` | Catatan |
|---|---|---|
| `GET /api/v1/web/notifications/inbox` | `NotificationItem[]` + `meta` | query: `unread_only`, `page`, `limit` (default limit 20, max 50). Meta: `current_page`, `per_page`, `total`, `total_pages`. |
| `GET /api/v1/web/notifications/unread-count` | `{ count: number }` | untuk badge |
| `POST /api/v1/web/notifications/:id/read` | `null` | `:id` = `NotificationItem.id` |
| `POST /api/v1/web/notifications/read-all` | `{ marked: number }` | tandai semua |
| `GET /api/v1/web/notifications/preferences` | `NotificationPreferenceResponse` | auto-create default bila belum ada |
| `PUT /api/v1/web/notifications/preferences` | `NotificationPreferenceResponse` | partial update (semua field pointer/opsional) |

### `NotificationItem`
```ts
{
  id: string          // delivery_attempt id → dipakai untuk POST /:id/read
  type: string        // system|social|content|reminder|security
  title: string
  body: string
  image_url?: string
  module: string      // identity|keuangan|psb|akademik|kesantrian|announcement
  event_type: string  // login_succeeded|payment_verified|...
  entity_id: string
  click_action: string  // route/deep link — boleh kosong di fase 1
  bypass: boolean
  extra?: Record<string, string>
  is_read: boolean
  created_at: string  // RFC3339
  read_at?: string
}
```

### `NotificationPreferenceResponse`
```ts
{
  id: string
  user_id: string
  all_notifications_enabled: boolean
  do_not_disturb_enabled: boolean
  do_not_disturb_start_time?: string  // "HH:MM"
  do_not_disturb_end_time?: string
  created_at: string
  updated_at: string
}
```

## Temuan penting dari kode backend (untuk UI)

1. **`meta` pagination inbox sekarang sudah dikembalikan.** Handler `ListInbox`
   mengembalikan `SuccessWithMeta` dengan `meta` (`current_page`, `per_page`,
   `total`, `total_pages`) sesuai konvensi sipon-be. UI bisa menggunakan
   `meta.total_pages` untuk navigasi atau tombol "Muat lagi". *(Sudah diperbaiki.)*
2. **`POST /:id/read` ownership check di SQL** — notifikasi milik user lain → `ERR_NOT_FOUND` (404).
3. **`read-all` vs `/:id/read`** — route static `read-all` didahulukan oleh router
   Gin, jadi `POST /read-all` aman tidak tertangkap param `:id`.
4. **Preferensi auto-create** — `GET /preferences` selalu mengembalikan row (default
   `all_enabled=true`, `dnd=false`). Tidak ada state "belum ada".
5. **DND validasi di backend** — saat `do_not_disturb=true`, `start_time` & `end_time`
   wajib ada (400 `ERR_BAD_REQUEST` bila kosong) dan format `HH:MM` valid (422
   `ERR_UNPROCESSABLE_ENTITY`). UI wajib enforce ini sebelum submit.
6. **`click_action` kosong di fase 1** — event `login_succeeded` belum set deep link,
   jadi item tidak navigasi kemana-mana. UI render link hanya bila `click_action`
   non-empty.
7. **`module`/`event_type`** bisa dipakai untuk ikon/warna per sumber (identity,
   keuangan, psb, dst.) — nilai sudah string, siap di-map ke ikon di client.

## Pola existing yang di-reuse

- `useApi()` / `parseApiError()` / envelope `ApiSuccess` (pola `app/stores/feedback.ts`).
- Store Pinia per domain (`app/stores/*.ts`).
- `usePermission()` untuk gating (hanya `member` auth; tanpa permission key di sini).
- `UDropdownMenu`, `UButton`, `UBadge`, `UIcon` (Nuxt UI) — pola `AppUserMenu.vue`.
- **Bell icon placeholder sudah ada** di `AppNavbar.vue` (baris `<UIcon name="i-lucide-bell" />`
  dalam `UButton ghost square`) — tinggal disambungkan ke dropdown + badge.
- Format tanggal/relatif: ikuti pola yang ada di `article`/`feedback` (`timeAgo` dsb.).

## Struktur baru

### Types — `shared/types/Notification.ts`
```typescript
export interface NotificationItem {
  id: string
  type: string
  title: string
  body: string
  image_url?: string
  module: string
  event_type: string
  entity_id: string
  click_action: string
  bypass: boolean
  extra?: Record<string, string>
  is_read: boolean
  created_at: string
  read_at?: string
}

export interface UnreadCountResponse { count: number }

export interface NotificationPreference {
  id: string
  user_id: string
  all_notifications_enabled: boolean
  do_not_disturb_enabled: boolean
  do_not_disturb_start_time?: string
  do_not_disturb_end_time?: string
  created_at: string
  updated_at: string
}

export interface UpdateNotificationPreferenceRequest {
  all_notifications_enabled?: boolean
  do_not_disturb_enabled?: boolean
  do_not_disturb_start_time?: string
  do_not_disturb_end_time?: string
}
```

### Store — `app/stores/notification.ts`
State: `items: NotificationItem[]`, `unreadCount: number`, `preference: NotificationPreference | null`,
`isLoading`, `isSubmitting`, `error`.

Actions:
- `fetchInbox(query)` → `GET /inbox` (`{ unread_only, page, limit }`); `append` bila `page > 1`.
- `fetchUnreadCount()` → `GET /unread-count` → set `unreadCount`.
- `markRead(id)` → `POST /:id/read`, lalu update item lokal (`is_read=true`) + `unreadCount` decrement (floor 0).
- `markAllRead()` → `POST /read-all`, lalu set semua `is_read=true` + `unreadCount=0`.
- `fetchPreference()` → `GET /preferences`.
- `updatePreference(payload)` → `PUT /preferences`, update `preference` dari response.

> Badge harus refresh saat app mount + setelah login. Taruh pemanggilan
> `fetchUnreadCount()` di layout (onMounted) atau plugin route, agar bell di
> semua halaman menampilkan angka benar.

### Komponen — `app/components/notification/`
- **`NotificationBell.vue`** — bungkus `UDropdownMenu` + `UButton` bell + `UBadge`
  (angka `unreadCount`, hidden bila 0). Saat dibuka → `fetchInbox({ unread_only: false, page: 1, limit: 10 })`.
  Render 5–10 item terbaru + footer link "Lihat semua" → `/notifikasi`.
- **`NotificationItemRow.vue`** — ikon per `module` (`i-lucide-bell` fallback), `title`
  (bold bila unread), `body` (truncate 2 baris), waktu relatif, tombol tandai dibaca
  (dot/action) bila unread; klik → `markRead(id)` + optional navigate(`click_action`) bila ada.
- **`NotificationPreferenceForm.vue`** — form preferensi: `UToggle`/`UCheckbox`
  `all_notifications_enabled`, toggle `do_not_disturb_enabled`, dua input waktu
  `HH:MM` (muncul hanya saat DND on), tombol simpan → `updatePreference`. Validasi
  client: saat DND on, kedua waktu wajib + format valid.

### Halaman
- **`app/pages/notifikasi/index.vue`** (layout `default`) — inbox penuh: filter
  `unread_only` (`UTabs`/`UButton` toggle), list `NotificationItemRow`, tombol
  "Tandai semua dibaca" (`markAllRead`), "Muat lagi" (`page++`).
- **`app/pages/notifikasi/pengaturan.vue`** — halaman preferensi, render
  `NotificationPreferenceForm` (fetch on mount). Entry point bisa dari:
  (a) menu dropdown user (`AppUserMenu`), atau (b) footer link di dropdown bell.

### Navigasi — wiring bell
- **`AppNavbar.vue`**: ganti tombol bell placeholder dengan `<NotificationBell />`.
- **`AppMobileBottomNav.vue`** (bila bell/notif relevan): opsional tambah item
  `Notifikasi` → `/notifikasi` (sesuaikan slot yang ada).
- Tambah link "Notifikasi" di menu user (`AppUserMenu.vue`) → `/notifikasi`, dan
  "Preferensi Notifikasi" → `/notifikasi/pengaturan` (pola "My Profile" yang ada).

## Backend — opsional (dokumentasi, bukan wajib fase ini)

1. **`click_action`** untuk event login/keuangan/psb belum diisi — diisi di fase
   lanjutan agar item bisa deep-link.

## Fase Pengerjaan

1. Types + store `notification.ts` (fetch inbox, unread count, mark read/all, preference).
2. Komponen `NotificationItemRow` + `NotificationBell`; wiring bell di `AppNavbar.vue`.
3. Halaman `/notifikasi` (inbox + filter unread + mark all + load more).
4. Komponen `NotificationPreferenceForm` + halaman `/notifikasi/pengaturan`.
5. Badge lifecycle (refresh on mount/login), link menu user.
6. Polish: empty state, error state, ikon per module, format waktu, responsive.

## Verifikasi

1. `npm run dev` (sipon-ui) + `go build` (sipon-be) lolos.
2. Login → bell muncul dengan badge `unread-count`; buka dropdown tampil notif
   `login_succeeded` terbaru (unread bold).
3. Klik item unread → `POST /:id/read`, badge berkurang; klik "Tandai semua" → badge 0.
4. Halaman `/notifikasi` list benar + filter `unread_only`; "Muat lagi" menambah item.
5. `/notifikasi/pengaturan`: matikan `all_notifications_enabled` → `PUT` sukses;
   aktifkan DND tanpa waktu → validasi client muncul (dan backend 400 bila lolos client).
6. Tidak ada endpoint yang butuh permission key — user `member` mendapat 200 (bukan 403).
