# Rules: Notification Mobile (In-App + Push)

Aturan teknis & bisnis yang wajib dipatuhi integrasi mobile (`sipon-app`) dengan
modul notifikasi backend (`sipon-be`). Ini melengkapi
`docs/specs/notification-mobile-api.md` (kontrak API/payload) dan
`docs/plan/notification-module.md` (desain modul).

---

## 1. Device Registration & Siklus Token

1. **Register setelah login sukses.** Aplikasi meminta FCM token lalu
   `POST /notifications/devices` setiap kali sesi baru dimulai (login) —
   jangan hanya saat pertama install.
2. **Unregister saat logout.** `DELETE /notifications/devices` dengan
   `provider_token`. Best-effort: tetap lanjutkan logout lokal walau panggilan
   gagal.
3. **Token per install, bukan per user.** `provider_token` unik; bila token
   berubah (reinstall / refresh token FCM), lakukan register ulang — backend
   menangani duplikasi (UNIQUE constraint upsert).
4. **Satu FCM token cukup untuk Android & iOS** — `push_provider` selalu `fcm`.
   Token FCM iOS dan Android sama-sama dikirim ke endpoint yang sama.
5. **Device nonaktif otomatis** saat backend menerima error token permanen
   (unregistered/not-registered/sender-id mismatch) — aplikasi tidak perlu
   menangani, cukup re-register bila FCM memberi token baru.

## 2. Sumber Notifikasi (In-App vs Push)

1. **In-app = read model DB.** Sumber kebenaran adalah
   `GET /notifications/inbox` + `GET /notifications/unread-count`. Push
   notification adalah **sinyal/duplikat** dari notifikasi yang sama — jangan
   jadikan push sebagai sumber kebenaran isi notifikasi.
2. **Sinkronisasi inbox via polling** saat app dibuka / on foreground (fase
   sekarang). Realtime (SSE/WS) **di luar cakupan** — lihat
   `docs/plan/notification-module.md` fase 3.
3. **Badge unread** diambil dari `GET /unread-count`; diperbarui saat: app
   mount, kembali ke foreground, dan setelah aksi mark-read.
4. **Mark read memakai delivery attempt id** (`NotificationItem.id`), bukan
   notification id. Ownership dicek backend (notif user lain → 404).

## 3. Payload Push & Perilaku OS

1. **`data` key kontrak** — aplikasi wajib membaca `module`, `event_type`,
   `entity_id`, `click_action`, `extra` persis sesuai spec. `extra` adalah
   string JSON — wajib di-parse, dan **tahan terhadap null/malformed** (pakai
   try-catch; gagal parse → abaikan, tetap boleh routing via `click_action`).
2. **Foreground** — FCM default tidak menampilkan tray notification saat app di
   depan. Aplikasi memilih: tampilkan local notification sendiri **atau** cukup
   refresh badge inbox + tampilkan in-app banner.
3. **Background/terminated** — OS menampilkan system tray (title+body).
   `click_action` baru tersedia saat user **mengetuk** notifikasi (bukan saat
   app terbuka). Data push tidak boleh menjadi satu-satunya jalur data;
   selalu verifikasi ke inbox bila butuh isi lengkap.
4. **Channel Android `sipon_notifications`** harus dibuat aplikasi (high
   importance) karena backend mengirim `android.notification.channel_id` ke
   channel tersebut. Tanpa channel ini di Android 8+, notifikasi tidak muncul.
5. **Notifikasi operasional (`bypass=true`)** berprioritas `high` untuk push.
   Namun **preferensi `all_notifications_enabled=false` tetap mematikan semua
   pengiriman** (dispatcher memeriksa preferensi sebelum membuat delivery),
   termasuk yang `bypass`.

## 4. Deep Link / Routing

1. **`click_action` kosong → buka inbox** (`/notifikasi`). Jangan crash.
2. **`click_action` tidak dikenal → buka inbox** (fallback aman).
3. **Path berisi id → ekstrak dari segmen**, lalu halaman tujuan membaca id
   dari route parameter (bukan dari `extra`).
4. **Routing dari push dan dari inbox memakai fungsi yang sama** — satu helper
   `handleClickAction(action, entityId)` agar perilaku konsisten.
5. **Auto mark-read saat navigasi**: ketika user membuka notifikasi (dari push
   maupun inbox), tandai dibaca via `POST /:id/read` (bila id delivery attempt
   tersedia) agar badge konsisten.

## 5. Preferensi & DND

1. **Auto-create di backend** — `GET /preferences` selalu mengembalikan row;
   aplikasi tidak perlu menangani state "belum ada".
2. **Enforce DND di sisi UI** — saat `do_not_disturb_enabled=true`, kedua waktu
   wajib terisi & format `HH:MM`. Backend memvalidasi ulang (400/422).
3. **DND belum di-enforce dispatcher saat ini** — backend hanya menyimpan
   preferensi (validasi format & kelengkapan waktu). Penerapan DND terhadap
   push/email/realtime adalah fase lanjutan. Aplikasi menampilkan status
   preferensi apa adanya dari backend.
4. **`all_notifications_enabled=false`** mematikan pengiriman (push **dan**
   in-app sama-sama tidak dibuat) karena dispatcher memeriksa preferensi
   sebelum menulis delivery. Aplikasi boleh tetap menampilkan inbox yang sudah
   ada.

## 6. Keamanan

1. Semua endpoint notifikasi butuh JWT bearer; **tidak ada endpoint notifikasi
   publik**.
2. Device registration self-scoped dari JWT — jangan pernah kirim `user_id` di
   body.
3. Jangan simpan `provider_token` / FCM token di storage yang tidak aman;
   cukup di memory provider saat runtime (token didapat ulang dari FCM bila
   perlu).
4. Logout harus menghapus token device di server agar user lama tidak lagi
   menerima push setelah logout.

## 7. Event yang Menghasilkan Notifikasi (acuan)

Event dipublish ke outbox oleh modul bisnis dan dikonsumsi worker notification.
Detail payload per event mengikuti handler `internal/modules/notification/
interfaces/mq/`. Mapping `click_action` dan `module`/`event_type` ada di
`docs/specs/notification-mobile-api.md` Bab 5.