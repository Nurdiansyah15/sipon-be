# API Spec: Notification Mobile (In-App + Push + Deep Link)

Dokumen ini adalah kontrak API dan payload push untuk integrasi **mobile**
(`sipon-app`, Flutter). Semua endpoint berada di bawah base path backend
`/api/v1/web` dan menggunakan **JWT bearer** (`jwtAuth` + `principalLoad`).

Envelope respons standar SIPON:

```json
{
  "status": "success",
  "status_code": 200,
  "message": "…",
  "data": { },
  "meta": null
}
```

---

## 1. Registrasi Device (untuk Push Notification)

Device harus **didaftarkan setelah login berhasil** agar backend bisa mengirim
push ke perangkat tersebut. Registrasi self-scoped: `user_id` diambil dari JWT,
bukan dari body.

### Register Device
```
POST /api/v1/web/notifications/devices
Authorization: Bearer <access_token>
```
Body:
```json
{
  "platform": "android",
  "push_provider": "fcm",
  "provider_token": "<FCM registration token>",
  "device_id": "optional-instance-id",
  "device_name": "POCO X3",
  "device_model": "M2010J19SY",
  "os_version": "14",
  "app_version": "0.1.0",
  "timezone": "Asia/Jakarta"
}
```
- `platform`: `android` | `ios` | `web` (default `android`).
- `push_provider`: `fcm` | `apns` | `huawei` | `web_push` (default `fcm`).
  Untuk Android & iOS saat ini selalu kirim `fcm` — token berasal dari
  Firebase Messaging yang menangani kedua platform.
- `provider_token`: token unik per install, **wajib**. `UNIQUE(provider_token)`,
  sehingga re-register token yang sama (mis. update app) tidak duplikat.

Respons `200`:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "device berhasil didaftarkan",
  "data": {
    "id": "…",
    "platform": "android",
    "push_provider": "fcm",
    "active": true,
    "last_seen_at": "2026-08-21T09:00:00+07:00"
  },
  "meta": null
}
```

### Unregister Device (Logout)
```
DELETE /api/v1/web/notifications/devices
Authorization: Bearer <access_token>
```
Body:
```json
{ "provider_token": "<FCM registration token>" }
```
- Dipanggil saat **logout** / akun dinonaktifkan agar device tidak lagi
  menerima push. Token tidak terdaftar untuk user → `404 ERR_NOT_FOUND`;
  aplikasi cukup **mengabaikan** 404 ini (best-effort).
- Respons `200` + `data: null`.

### List Device (Opsional)
```
GET /api/v1/web/notifications/devices
Authorization: Bearer <access_token>
```
Respons `200` + `data: DeviceResponse[]`.

---

## 2. Inbox (Notifikasi In-App)

Inbox adalah riwayat notifikasi per user yang sudah dibaca/ditandai dibaca.

### List Inbox
```
GET /api/v1/web/notifications/inbox?unread_only=true&page=1&limit=20
Authorization: Bearer <access_token>
```
Query:
- `unread_only` (bool, default `false`)
- `page` (default `1`)
- `limit` (default `20`, maks `50`)

Respons `200`, `data: NotificationItem[]`, `meta`:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "inbox notifikasi berhasil diambil",
  "data": [
    {
      "id": "<delivery_attempt_id>",
      "type": "system",
      "title": "Kehadiran Tercatat",
      "body": "Selamat Ahmad, kehadiran Anda berhasil tercatat.",
      "image_url": "",
      "module": "akademik",
      "event_type": "attendance_recorded",
      "entity_id": "<attendance_id>",
      "click_action": "/akademik/absensi",
      "bypass": true,
      "extra": { "session_id": "…", "source": "fingerprint" },
      "is_read": false,
      "created_at": "2026-08-21T09:00:00+07:00",
      "read_at": null
    }
  ],
  "meta": {
    "current_page": 1,
    "per_page": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

Catatan untuk mobile:
- `id` adalah **delivery attempt id** — dipakai untuk `POST /:id/read`.
- `click_action` berisi path tujuan (web route). Lihat **Bab 5 (Deep Link)**
  untuk pemetaan ke route `sipon-app`.
- `is_read` dipakai untuk styling (bold/unbold); `read_at` ada bila sudah dibaca.
- `extra` map string opsional, bisa berisi konteks tambahan per event.

### Unread Count (Badge)
```
GET /api/v1/web/notifications/unread-count
Authorization: Bearer <access_token>
```
Respons `200`:
```json
{ "data": { "count": 3 }, "status": "success", "status_code": 200, "message": "…", "meta": null }
```

### Tandai Satu Dibaca
```
POST /api/v1/web/notifications/:id/read
Authorization: Bearer <access_token>
```
- `:id` = `NotificationItem.id` (delivery attempt id).
- Ownership di-cek di SQL; notifikasi milik user lain → `404 ERR_NOT_FOUND`.
- Respons `200` + `data: null`.

### Tandai Semua Dibaca
```
POST /api/v1/web/notifications/read-all
Authorization: Bearer <access_token>
```
Respons `200`:
```json
{ "data": { "marked": 3 }, "status": "success", "status_code": 200, "message": "…", "meta": null }
```

---

## 3. Preferensi Notifikasi

Preferensi di-*auto-create* saat pertama diakses (default semua aktif, DND off).

### Get Preferensi
```
GET /api/v1/web/notifications/preferences
Authorization: Bearer <access_token>
```
Respons `200`:
```json
{
  "data": {
    "id": "…",
    "user_id": "…",
    "all_notifications_enabled": true,
    "do_not_disturb_enabled": false,
    "do_not_disturb_start_time": null,
    "do_not_disturb_end_time": null,
    "created_at": "…",
    "updated_at": "…"
  },
  "status": "success",
  "status_code": 200,
  "message": "…",
  "meta": null
}
```

### Update Preferensi
```
PUT /api/v1/web/notifications/preferences
Authorization: Bearer <access_token>
```
Body (semua field opsional / partial):
```json
{
  "all_notifications_enabled": false,
  "do_not_disturb_enabled": true,
  "do_not_disturb_start_time": "22:00",
  "do_not_disturb_end_time": "07:00"
}
```
Validasi backend:
- Saat `do_not_disturb_enabled = true`, `start_time` & `end_time` **wajib ada**
  (jika kosong → `400`).
- Format waktu `HH:MM` tidak valid → `422 ERR_UNPROCESSABLE_ENTITY`.
- Respons `200` + `NotificationPreferenceResponse`.

---

## 4. Payload Push Notification (FCM)

Backend mengirim push lewat Firebase Cloud Messaging (`firebase_messaging`).
Struktur FCM message:

- **`notification.title`** — judul (muncul di system tray).
- **`notification.body`** — isi (muncul di system tray).
- **`data`** — payload string (dikirim ke handler, berguna saat app dibuka
  dari notifikasi).

```jsonc
{
  "notification": { "title": "Kehadiran Tercatat", "body": "…" },
  "data": {
    "module": "akademik",
    "event_type": "attendance_recorded",
    "entity_id": "<attendance_id>",
    "click_action": "/akademik/absensi",
    "extra": "{\"session_id\":\"…\",\"source\":\"fingerprint\"}"
  },
  "android": {
    "priority": "high",
    "notification": { "channel_id": "sipon_notifications" }
  },
  "apns": {
    "headers": { "apns-priority": "10" },
    "payload": { "aps": { "sound": "default", "content-available": true, "badge": 1 } }
  }
}
```

### Kontrak key `data` (WAJIB dipatuhi aplikasi)

| Key | Tipe | Selalu ada | Keterangan |
|---|---|---|---|
| `module` | string | ya | `identity`, `keuangan`, `psb`, `article`, `akademik`, `announcement` |
| `event_type` | string | ya | contoh `attendance_recorded`, `payment_verified` |
| `entity_id` | string | ya | id entitas terkait, bisa kosong |
| `click_action` | string | ya (bisa `""`) | path tujuan; kosong → buka inbox |
| `extra` | string (JSON) | opsional | JSON-encode dari `map[string]string`; parse dengan `jsonDecode` |

### Prioritas & preferensi
- Notifikasi **operasional** (`bypass = true`, contoh kehadiran & pembayaran)
  dikirim dengan `priority=high` — muncul langsung.
- `bypass` **saat ini hanya menaikkan prioritas push**; preferensi
  `all_notifications_enabled` tetap dicek dispatcher (bila `false`, push dan
  in-app sama-sama tidak dibuat). DND belum di-enforce di fase ini.

### Perilaku sisi aplikasi
- App **foreground**: tampilkan sebagai local notification (atau update badge
  inbox saja sesuai kebijakan UX).
- App **background / terminated**: tampilkan system tray (dilakukan OS);
  saat user mengetuk notifikasi, app dibuka dan handler menerima `data` untuk
  routing (lihat Bab 5).

---

## 5. Deep Link / Routing ke Halaman

Sumber routing ada dua:

1. **Dari push notification** (getInitialMessage / onMessageOpenedApp) — key
   `data.click_action`.
2. **Dari inbox in-app** (ketuk item) — field `click_action` pada
   `NotificationItem`.

### Kontrak `click_action`

`click_action` adalah **path web (web route)** yang dikirim backend. Aplikasi
memetakannya ke route `go_router`-nya sendiri. Kontraknya:

- Selalu diawali `/` (root-relative path web).
- `""` (kosong) → fallback ke halaman inbox notifikasi.
- Path dapat memuat segmen id, contoh `/keuangan/tagihan/<invoice_id>`,
  `/artikel/<article_id>`. Aplikasi harus mengekstrak id dari segmen.
- Bila path tidak dikenal aplikasi → fallback ke inbox (jangan crash).

### Daftar `click_action` yang diterbitkan backend

| Modul | Event | `click_action` | Parameter |
|---|---|---|---|
| identity | `login_succeeded` | *(kosong)* | — |
| psb | `pendaftaran_submitted`, `daftar_ulang_submitted`, `dokumen_verified`, `dokumen_rejected`, `revision_requested`, `revision_requested_daftar_ulang`, `pendaftaran_accepted`, `pendaftaran_rejected`, `nis_generated` | `/psb/riwayat` | — |
| article | `article_published` | `/artikel/<id>` | `entity_id` |
| article | `articles_scraped` | `/artikel` | — |
| keuangan | `invoice_issued`, `invoice_cancelled`, `payment_submitted`, `payment_verified`, `payment_rejected` | `/keuangan/tagihan/<id>` | `entity_id` |
| akademik | `session_reminder` | `/akademik` | — |
| akademik | `attendance_recorded` | `/akademik/absensi` | — |

### Error response (HTTP)

`httperror` mengembalikan body berikut (pola yang sama untuk semua endpoint):

```json
{
  "status": "error",
  "status_code": 401,
  "error_code": "ERR_UNAUTHORIZED",
  "errors": "…"
}
```

Error yang mungkin relevan untuk mobile: `401 ERR_UNAUTHORIZED` (token habis —
refresh sekali via `POST /web/auth/refresh-token`, lalu logout bila gagal),
`404 ERR_NOT_FOUND`, `422 ERR_UNPROCESSABLE_ENTITY`, `409 ERR_CONFLICT`,
`400 ERR_BAD_REQUEST`.