# Plan: Integrasi Notifikasi Mobile (`sipon-app` + `sipon-be`)

## Context

Backend modul `notification` sudah berjalan (lihat `docs/plan/notification-module.md`):
in-app inbox, preferensi, **device registration**, dan **push via Firebase Cloud
Messaging (FCM)** sudah diimplementasikan di `internal/modules/notification/`.
Web UI (`sipon-ui`) juga sudah direncanakan (`docs/plan/notification-ui.md`).

Yang belum ada: **integrasi aplikasi mobile** (`sipon-app`, Flutter) untuk
menerima notifikasi. Saat ini `sipon-app` hanya punya halaman login/register dan
dashboard, serta belum punya kemampuan push / inbox / deep-link.

Dokumen ini menjabarkan rencana integrasi di **dua repositori**:

- **`sipon-be`** — penyempurnaan backend agar kontrak push & deep-link siap
  dikonsumsi mobile.
- **`sipon-app`** — implementasi Flutter: FCM, registrasi device, inbox,
  preferensi, badge, dan **routing notifikasi ke halaman yang ditentukan**.

Kontrak API & payload dirinci di `docs/specs/notification-mobile-api.md`; aturan
teknis/bisnis di `docs/rules/notification-mobile.md`.

## Tujuan & Cakupan

1. `sipon-app` menerima **push notification** (foreground, background,
   terminated) dari FCM.
2. `sipon-app` menampilkan **inbox notifikasi in-app** (list, badge unread,
   mark read, mark all read) dan **preferensi** (global on/off + DND).
3. `sipon-app` **merouting notifikasi ke halaman yang ditentukan** oleh
   `click_action` — dari ketukan push maupun ketukan item inbox.
4. `sipon-be` memastikan payload push & `click_action` konsisten dan siap
   dikonsumsi aplikasi.

**Cakupan berakhir di poin 3** — sampai aplikasi bisa membuka halaman tujuan
yang benar dari notifikasi. Realtime (SSE/WS) dan channel email/SMS **di luar
cakupan** (lihat fase lanjutan `docs/plan/notification-module.md`).

## Kondisi Saat Ini (audit)

### `sipon-be` — sudah ada ✓

| Kapabilitas | Lokasi |
|---|---|
| Tabel `notifications`, `delivery_attempts`, `notification_preferences`, `device_registrations` | `migrations/20260819*.up.sql` |
| Endpoint inbox / unread / read / read-all / preferences | `internal/modules/notification/interfaces/http/` |
| Endpoint device: `POST/DELETE/GET /devices` | `internal/modules/notification/application/command/device.go` |
| Push sender FCM (priority, channel `sipon_notifications`, APNs) | `internal/modules/notification/infrastructure/external/fcm_sender.go` |
| Dispatcher unicast/multicast/broadcast + push data (`module`, `event_type`, `entity_id`, `click_action`, `extra`) | `internal/modules/notification/application/dispatcher.go` |
| Event-driven via outbox + RabbitMQ (`sipon.worker.notification`) | `internal/modules/notification/interfaces/mq/` |
| `click_action` pada sebagian event (psb, keuangan, artikel, akademik) | `internal/modules/notification/interfaces/mq/handler.go` |
| Konfig FCM (`FCM_CREDENTIALS_PATH`, `FCM_PROJECT_ID`) | `internal/shared/config/config.go` |

### `sipon-app` — saat ini ✗

- Flutter 3.9 / `go_router` 17 / `dio` / `provider` / `shared_preferences`.
- Hanya route `/login`, `/register`, `/dashboard` (`lib/shared/router/app_router.dart`).
- Belum ada: `firebase_messaging`, config FCM (google-services/entitlements),
  store/provider notifikasi, halaman inbox/preferensi, deep-link handler.

## Arsitektur Alur

### Alur push (end-to-end)

```
Event bisnis (akademik.attendance.recorded, dst.)
  │  publish ke event_outbox (dalam transaksi bisnis)
  ▼
Outbox Relay (worker) → RabbitMQ → Message Consumer
  ▼
notification handler (worker) → Dispatcher.Dispatch (channel in_app + push)
  ├─► delivery_attempts (in_app)  → dibaca app via GET /inbox + unread-count
  └─► FCM Send (title/body + data {module,event_type,entity_id,click_action,delivery_id,extra})
       └─► device_registrations[provider_token]  ──►  OS tray (bg/terminated)
                                                        ▼ (user tap)
                                              app handle data.click_action
                                                        ▼
                                              go_router push → halaman tujuan
```

### Alur in-app (manual)

```
App buka / app foreground → GET /inbox + GET /unread-count (badge)
  → user ketuk item → mark read + navigate(click_action) → halaman tujuan
```

## Rencana `sipon-be` (penyempurnaan kecil)

> **Status: SELESAI** — bagian backend untuk integrasi mobile sudah
> diimplementasikan di `sipon-be`. Sisa pekerjaan hanya verifikasi bersama
> aplikasi (smoke test perangkat nyata) yang menunggu sisi `sipon-app`.

1. ✅ **Audit & lengkapi `click_action`** — semua event notifikasi sudah punya
   `click_action` (lihat tabel di spec Bab 5), kecuali
   `identity.user.login_succeeded` yang sengaja kosong (fallback inbox).
2. ✅ **`data` push menambah `delivery_id`** — `Dispatcher.buildPushMessages`
   menambahkan key `delivery_id` (delivery attempt id) ke payload push agar app
   bisa langsung mark-read saat notifikasi diketuk tanpa fetch inbox.
   `internal/modules/notification/application/dispatcher.go`.
3. ✅ **Perbaiki badge APNs** — badge APNs tidak lagi statis `1`; sekarang
   dihitung dari unread inbox per user (`CountUnreadInApp`) dan dikirim via
   `PushMessage.UnreadCount`. `internal/modules/notification/infrastructure/external/fcm_sender.go`.
4. ✅ **Worker mendeklarasikan binding notifikasi** — sudah ada sejak awal
   (`cmd/worker/main.go` register `notification`).
5. ⏳ **Uji dengan aplikasi nyata (smoke test)** — kirim broadcast/push ke token
   dev, verifikasi title/body/data sampai perangkat. **Menunggu implementasi
   `sipon-app`** (lihat Verifikasi).

> Backend TIDAK perlu menambahkan endpoint baru untuk cakupan ini. Semua yang
> dibutuhkan app sudah tersedia: device, inbox, unread-count, read, preferensi.

## Rencana `sipon-app` (Flutter)

### 0. Prasyarat Firebase

- Buat project Firebase; tambahkan app **Android** (`google-services.json`) dan
  **iOS** (`GoogleService-Info.plist`) untuk bundle `sipon_app`.
- Dependensi baru di `pubspec.yaml`:
  - `firebase_core`
  - `firebase_messaging`
- Android: letakkan `google-services.json` di `android/app/`; tambahkan plugin
  `com.google.gms.google-services` di `android/settings.gradle` & `build.gradle`.
- iOS: `GoogleService-Info.plist` di Xcode; aktifkan Push Notifications +
  Background Modes (Remote notifications) di capabilities; panggil
  `FirebaseMessaging.instance.setForegroundNotificationPresentationOptions`.
- Konfig backend `.env` (API): pastikan `FCM_CREDENTIALS_PATH` & `FCM_PROJECT_ID`
  terisi di `sipon-be` agar push aktif.

### 1. Inisialisasi & izin

- `main.dart`: init `Firebase.initializeApp()` sebelum `runApp`.
- Minta izin notifikasi saat login sukses / pertama masuk
  (`FirebaseMessaging.instance.requestPermission()`).
- Buat channel Android `sipon_notifications` (high importance) — wajib agar
  push backend muncul (backend mengirim `channel_id` ini).

### 2. Registrasi device (siklus sesi)

- **Login sukses** → ambil `FirebaseMessaging.instance.getToken()` →
  `POST /notifications/devices` (platform & push_provider `fcm`).
- Simpan `providerToken` di memory provider (bukan prefs kalau bisa dihindari).
- **Logout** → `DELETE /notifications/devices` (best-effort).
- Listen `FirebaseMessaging.instance.onTokenRefresh` → re-register token baru.

### 3. Store & provider notifikasi (pola yang ada)

Ikuti pola feature saat ini (`lib/features/*`):

- `features/notification/domain/entities/notification_item.dart`
  (model `NotificationItem`, `UnreadCount`, `NotificationPreference`).
- `features/notification/data/datasources/notification_remote_data_source.dart`
  (dio, base URL dari `ApiConstants`).
- `features/notification/data/repositories/notification_repository_impl.dart`.
- `features/notification/presentation/providers/notification_provider.dart`
  (ChangeNotifier + `provider`): `items`, `unreadCount`, `preference`,
  actions `fetchInbox`, `fetchUnreadCount`, `markRead`, `markAllRead`,
  `fetchPreference`, `updatePreference`.

### 4. Screens

- **Inbox** `features/notification/presentation/screens/inbox_screen.dart`
  (route `/notifikasi`): list `NotificationItem`, filter `unread_only`,
  tombol "tandai semua", "muat lagi", pull-to-refresh, empty state.
- **Preferensi** `.../notification_preference_screen.dart` (route
  `/notifikasi/pengaturan`): toggle global, toggle DND + 2 input waktu
  (enforce waktu wajib saat DND on, format `HH:MM`).
- **Badge unread** pada area navigasi/dashboard: angka dari
  `unreadCount`, refresh saat app mount & kembali ke foreground.

### 5. Deep-link routing (inti cakupan)

Semua routing notifikasi lewat **satu helper**:

```dart
/// features/notification/presentation/notification_router.dart
void routeFromNotification(
  BuildContext context, {
  required String clickAction,
  String entityId = '',
  String? deliveryId, // bila tersedia → auto mark-read
}) {
  final router = GoRouter.of(context);
  switch (clickAction) {
    case '' : router.go('/notifikasi'); break;
    case '/akademik/absensi': router.go('/notifikasi'); break; // sampai halaman absensi ada
    case '/akademik': router.go('/dashboard'); break;          // sampai halaman akademik ada
    case '/psb/riwayat': router.go('/notifikasi'); break;      // sampai halaman psb ada
    default:
      if (clickAction.startsWith('/artikel/')) {
        router.push('/artikel/${clickAction.split('/').last}');
      } else if (clickAction.startsWith('/keuangan/tagihan/')) {
        router.push('/keuangan/tagihan/${clickAction.split('/').last}');
      } else {
        router.go('/notifikasi'); // fallback aman
      }
  }
  if (deliveryId != null) unreadProvider.markRead(deliveryId);
}
```

> **Fase pengerjaan**: halaman yang belum ada (artikel detail, tagihan detail,
> psb riwayat, absensi) dirender sebagai **stub** dengan tombol kembali ke
> inbox, sampai fitur masing-masing dikerjakan. Pemetaan final:

| `click_action` (backend) | Route `go_router` (target) | Status |
|---|---|---|
| `""` | `/notifikasi` | baru (fase ini) |
| `/akademik/absensi` | `/akademik/absensi` | stub fase ini |
| `/akademik` | `/dashboard` | sementara |
| `/psb/riwayat` | `/psb/riwayat` | stub fase ini |
| `/artikel/<id>` | `/artikel/<id>` | stub fase ini |
| `/artikel` | `/artikel` | stub fase ini |
| `/keuangan/tagihan/<id>` | `/keuangan/tagihan/<id>` | stub fase ini |

### 6. FCM handler

- `FirebaseMessaging.onMessageOpenedApp` (background → tap) &
  `FirebaseMessaging.instance.getInitialMessage()` (terminated → tap): parse
  `message.data`, panggil `routeFromNotification`.
- `FirebaseMessaging.onMessage` (foreground): kebijakan fase ini — tampilkan
  **local notification** sendiri (via `flutter_local_notifications`, optional)
  **atau** cukup refresh badge & tampilkan banner in-app; lalu
  `fetchUnreadCount()`.
- `AppLifecycleListener`/`WidgetsBindingObserver`: saat kembali ke foreground →
  refresh `unreadCount` + inbox.

### 7. Wiring DI

- `lib/core/di/app_providers.dart`: tambahkan `NotificationProviders.providers`
  (remote data source → repository → provider), dan pasang `navigatorKey`/
  `routeFromNotification` yang bisa dipanggil dari luar widget tree (mis.
  disimpan di `NotificationRouter` singleton yang membaca `GoRouter` dari
  context global / `scaffoldMessengerKey`).

## Fase Pengerjaan

1. ✅ **Fase 0 — Fondasi backend (sipon-be)**: audit `click_action`,
   `delivery_id` di data push, badge APNs dinamis, verifikasi FCM config &
   worker binding. **Selesai.** Bukti: `go build`/`go vet`/`go test` lolos.
2. ⏳ **Fase 1 — Fondasi app**: Firebase setup (core + messaging, config files),
   init di `main.dart`, izin notifikasi, channel Android, token flow
   (register device saat login, unregister saat logout).
3. ⏳ **Fase 2 — Inbox & badge**: model/repo/provider, halaman inbox, unread
   badge, mark read/all, preferensi (halaman + validasi DND).
4. ⏳ **Fase 3 — Deep-link routing (cakupan akhir)**: helper `routeFromNotification`,
   FCM foreground/background/terminated handler, stub halaman tujuan,
   auto mark-read saat navigasi.
5. ⏳ **Fase 4 — Polish & E2E**: pull-to-refresh, empty/error state, lokal notif
   foreground (opsional), test perangkat nyata.

## Verifikasi

### `sipon-be`
1. ✅ `go build ./...`, `go vet ./...`, `go test ./internal/modules/...` lolos.
2. ⏳ Broadcast admin ke token dev → push muncul dengan `data.click_action` &
   `data.delivery_id` benar. *(Butuh device/app — dilakukan bersama `sipon-app`.)*
3. ⏳ Event `akademik.attendance.recorded` → push + inbox item `click_action`
   `/akademik/absensi`. *(Butuh device/app.)*

### `sipon-app`
1. `flutter pub get`, `flutter analyze` bersih.
2. Login → device terdaftar (`GET /notifications/devices` berisi token app).
3. Push saat **background/terminated** → ketuk notifikasi → app terbuka dan
   route ke halaman sesuai `click_action`.
4. Push saat **foreground** → badge unread bertambah (atau local notif tampil
   sesuai kebijakan).
5. Inbox: list benar, badge berubah setelah mark-read/all, filter unread,
   preferensi tersimpan (DND tanpa waktu → error validasi UI).
6. Logout → device tidak lagi menerima push.

## Out of Scope

- Realtime (SSE/WebSocket) — fase 3 `docs/plan/notification-module.md`.
- Channel `email`/`sms` — fase 2 modul notification.
- Preferensi per-modul (keuangan/psb/akademik/…) — fase 2.
- Implementasi penuh halaman artikel/tagihan/psb/absensi — di sini hanya stub
  agar routing berfungsi.