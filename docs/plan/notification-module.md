# Plan: Module Notification (Notifikasi)

## Context

SIPON belum punya mekanisme notifikasi terpusat. Kejadian bisnis yang penting —
login berhasil, pembayaran terverifikasi/ditolak (`keuangan`), hasil review PSB
(`psb`), review dokumen santri (`kesantrian`), surat dibuat (`kesantrian/persuratan`),
dan pengumuman admin — hari ini tidak memberi tahu penggunanya.

Blueprint diambil dari implementasi notifikasi yang sudah matang di
`k-forum-ws/k-forum-api` (domain `notification`, service `dispatcher`, usecase
`notification`/`notificationpreference`, tabel `notifications`/`delivery_attempts`/`notification_preferences`/`device_registrations`),
lalu **diadaptasi** ke arsitektur modular monolith DDD milik SIPON
(`internal/modules/*`, lihat `docs/architecture/module-boundaries.md`).

Blueprint k-forum-api tidak disalin mentah-mentah karena perbedaan penting:

| Aspek | k-forum-api | SIPON (target) |
|---|---|---|
| Struktur folder | `internal/domain/*`, `internal/app/usecase/*`, `internal/interfaces/*` | `internal/modules/notification/{domain,application,infrastructure,interfaces}` |
| Error | `internal/domain/errors` (`domainerr`) | `internal/shared/kernel` (`AppError`, `New/Wrap`) |
| Error mapping | `apperror.*` | `application/errors.go` (kode `ERR_*` + `httperror.Handle`) |
| Trigger | MQ handler → dispatcher | `messaging` (outbox + RabbitMQ + registry) — **event-driven** |
| Async infra | outbox relay + worker | `event_outbox` + Outbox Relay + Message Consumer (sudah ada) |
| ID | `uuid.NewString()` (Go) | `gen_random_uuid()` DB default; `google/uuid` tersedia |
| User lookup | `userRepo.FindByID` (domain user) | `identity.Contract.GetUserSummary` (via `identitygateway`) |
| Realtime in-app | WebSocket hub | **belum ada** — pakai polling inbox (`list` + `unread-count`) |
| Push (FCM) | ya (device registration) | **belum ada** — fase lanjutan |

> **Skema yang TIDAK dipakai dari blueprint.** k-forum-api punya
> `notification_send_logs` + `notification_send_log_results` dan usecase
> `notificationtester` (riwayat "test kirim" dari backoffice). Skema ini **tidak
> berguna** — model blueprint + `delivery_attempts` sudah cukup menggantikannya.
> Test kirim ya test saja, **tidak perlu disimpan riwayatnya**. Karena itu send
> log/tester **tidak dibawa** ke SIPON.

## Tujuan & Cakupan

Fase 1 (MVP) — **in-app notification + event-driven**:
1. **Inbox** — list, unread count, mark read (satu & semua).
2. **Preferensi** — global on/off + Do-Not-Disturb (DND) per user.
3. **Trigger via event** — module bisnis mempublish domain event ke `event_outbox`;
   notification mengkonsumsinya. Contoh pertama: **`identity.user.login_succeeded`**.
4. **Broadcast admin** — kirim pengumuman ke banyak user (tanpa riwayat/send log).

Fase 2 — channel tambahan & preferensi per-modul:
5. Channel `email` / `sms` (reuse infra identity), channel `push` + `device_registrations`.
6. Preferensi per-modul SIPON (keuangan, psb, akademik, kesantrian, pengumuman).

Fase 3 — realtime:
7. Realtime (SSE/WebSocket) bila dibutuhkan client.

## Keputusan Arsitektur

1. **Satu module `notification`** dengan sub-aggregate domain:
   `notification` (blueprint), `delivery` (DeliveryAttempt), `preference`, dan
   (fase 2) `device`.
2. **Blueprint + DeliveryAttempt dipisah** (pola k-forum-api): `notifications`
   menyimpan intent/blueprint; `delivery_attempts` menyimpan notifikasi nyata
   per `user × channel` (1 baris = 1 penerima × 1 channel). Inbox = read model
   dari JOIN `delivery_attempts` (channel `in_app`) + `notifications`. Ini model
   yang **menggantikan** send-log/tester k-forum-api.
3. **Trigger utama = event-driven (asynchronous).** Module bisnis mempublish
   domain event ke `event_outbox` (dalam transaksi bisnisnya); notification
   mendaftarkan handler via `RegisterMessageHandlers` (pola `akademik/interfaces/mq`).
   Tidak ada `Contract.Send` sinkron dulu — YAGNI (lihat `module-boundaries.md`);
   ditambah bila ada module yang butuh kirim sinkron.
4. **In-app = tulis DB.** Tanpa hub realtime, "kirim in-app" cukup menulis
   `DeliveryAttempt` berstatus `success`; client membaca lewat endpoint inbox
   (polling) + `unread-count` untuk badge.
5. **Resolusi identitas via `identity.Contract`** — untuk channel non-in_app
   (fase 2) butuh email/phone, didapat dari `GetUserSummary`. Fase 1 (in-app)
   tidak butuh lookup identitas sama sekali.
6. **Tanpa FK ke tabel module lain** — `user_id` plain UUID; hanya `users(id)`
   yang di-FK untuk `delivery_attempts`/`notification_preferences`.
7. **Error mapping dua tingkat** — domain kembalikan `kernel.AppError` kode domain
   (`NOTIFICATION_NOT_FOUND`, …); `application/errors.go` memetakan ke `ERR_*`.
8. **ID di-generate di Go** (`google/uuid`) untuk blueprint & delivery attempt —
   dispatcher butuh ID sebelum persist; konsisten dengan messaging.

## Alur Trigger Event (contoh: login berhasil)

```
LoginUseCase (identity, API process)
  │  user.Update + outbox.Save("identity.user.login_succeeded", {"user_id": ...})
  │  (satu DB transaction — pola outbox, lihat scheduler)
  ▼
event_outbox (tabel DB)
  │  Outbox Relay (worker) claim → publish ke RabbitMQ
  ▼
Message Consumer (worker) → registry dispatch "identity.user.login_succeeded"
  ▼
notification handler (interfaces/mq) → dispatcher.Dispatch (in_app)
  ▼
delivery_attempts (inbox user) → dibaca client via GET /notifications/inbox
```

Perubahan yang diperlukan di sisi **identity**:
- `application/ports/outbox.go` — port `OutboxWriter{ Save(ctx, routingKey string,
  payload json.RawMessage) error }` (bentuk sama dengan `scheduler/application/ports/outbox.go`).
- `LoginUseCase` menulis event `identity.user.login_succeeded` di transaksi yang
  sama dengan `user.Update` (reset failed attempts). **Penting:** pakai
  `database.Transactor` (shared `internal/shared/database`) — BUKAN
  `persistence.PostgresTransactor` milik identity — supaya `txKey`-nya sama
  dengan yang dicek `messaging` outbox repo (`database.ExecerFromContext`), jadi
  outbox ikut commit/rollback bersama perubahan user.
- `module.go` mengekspos setter/param untuk memasang `OutboxWriter` (pola
  `scheduler.WithOutbox`/`kesantrian.SetAkademikProvisioner`).

Perubahan di sisi **composition root**:
- `cmd/api/main.go`: buat `outboxPersistence.NewPostgresOutboxRepository(db)`,
  bungkus jadi `identity.OutboxWriter` adapter, pasang ke identity. (API proses
  cuma menulis ke tabel `event_outbox`; relay/publish tetap di worker.)
- `cmd/worker/main.go`: tambah `register("notification", notification.RegisterMessageHandlers)`.

## Domain Model

### Constant (`domain/*/constant`)

`notification_constant.go` (kode error domain sekalian, pola `feedback`):
- `NotificationType`: `system`, `social`, `content`, `reminder`, `security`.
- `AudienceType`: `unicast`, `multicast`, `broadcast` (+ `IsValid()`).
- `NotificationChannel`: `in_app`, `email`, `sms`, `push` (+ `IsValid()`).
- `DeliveryStatus`: `pending`, `success`, `failed`, `retrying`.
- Kode error domain: `NOTIFICATION_NOT_FOUND`, `NOTIFICATION_TITLE_REQUIRED`,
  `NOTIFICATION_BODY_REQUIRED`, `NOTIFICATION_AUDIENCE_TYPE_INVALID`,
  `DELIVERY_NOT_FOUND`, `DELIVERY_PERSISTENCE_FAILED`, `PREFERENCE_*`, dst.

`device_constant.go` (fase 2): `Platform` (`android`/`ios`/`web`), `PushProvider`
(`fcm`/`apns`/`web_push`).

### Entity

**`notification`** — blueprint sebelum didistribusikan:
```go
type Notification struct {
    ID            string
    AudienceType  constant.AudienceType
    AudienceData  map[string]any   // {"user_ids": [...]} untuk unicast/multicast
    Type          constant.NotificationType
    Title         string
    Body          string
    Payload       valueobject.NotificationPayload
    ReferenceID   *string
    ReferenceType *string
    CreatedAt     time.Time
}
func NewNotification(params NotificationParams) (*Notification, error)
```

**`delivery`** — notifikasi nyata per `user × channel` (inbox):
```go
type DeliveryAttempt struct {
    ID             string
    NotificationID string
    UserID         string
    Channel        constant.NotificationChannel
    Status         constant.DeliveryStatus
    ProviderCode   *string
    RetryCount     int
    NextRetryAt    *time.Time
    AttemptedAt    time.Time
    ReadAt         *time.Time
}
// NewDeliveryAttempt(...), MarkSuccess(), MarkFailed(code), ScheduleRetry(at),
// IsRetryable(), MarkRead(), IsUnread()
```

**`preference`** — per user:
```go
type NotificationPreference struct {
    ID                   string
    UserID               string
    AllEnabled           bool
    DoNotDisturbEnabled  bool
    DNDStartTime         *string  // "HH:MM"
    DNDEndTime           *string
    ModulePreferences    ModulePreferences // JSONB (fase 2)
    CreatedAt, UpdatedAt time.Time
}
// NewNotificationPreference(id, userID) — default semua enabled
// IsWithinDND(now) bool — dukung rentang lewat tengah malam
// (fase 2) IsAllowedByModule(eventKey string) bool
```

### Value Object

**`valueobject/notification_vo.go`** — `NotificationPayload` (immutable):
```go
type NotificationPayload struct {
    Module      string            // "identity", "keuangan", "psb", "akademik", "kesantrian", "announcement"
    EventType   string            // "login_succeeded", "payment_verified", "review_approved", dst.
    EntityID    string
    ClickAction string            // deep link / route tujuan client
    ImageURL    string
    Bypass      bool
    Extra       map[string]string
}
```
`DoNotDisturbWindow{StartTime, EndTime}` + `NewDoNotDisturbWindow` (validasi HH:MM)
+ `IsActive(now)` (dukung overnight).

### Repository (`domain/*/repository/interfaces.go`)

```go
type NotificationRepository interface {
    Save(ctx, *entity.Notification) error
    FindByID(ctx, id string) (*entity.Notification, error)
}

type DeliveryAttemptRepository interface {
    Save(ctx, *entity.DeliveryAttempt) error
    FindByID(ctx, id string) (*entity.DeliveryAttempt, error)
    // inbox (channel in_app) — read model JOIN notifications
    ListInApp(ctx, q ListInAppQuery) ([]InboxReadItem, meta, error)
    CountUnreadInApp(ctx, userID string) (int64, error)
    MarkRead(ctx, id, userID string) error           // ownership di SQL
    MarkAllRead(ctx, userID string) (int, error)
    FindPendingRetries(ctx) ([]*entity.DeliveryAttempt, error) // fase lanjutan
}

type NotificationPreferenceRepository interface {
    FindOrCreateByUserID(ctx, userID string) (*entity.NotificationPreference, error)
    Update(ctx, *entity.NotificationPreference) error
    FindByUserIDs(ctx, userIDs []string) (map[string]*entity.NotificationPreference, error) // fase lanjutan (fanout bulk)
}

// fase 2
type DeviceRegistrationRepository interface { ... }
```

## Skema Data (migration `2026..._create_notification_tables`)

Mengikuti gaya SIPON (`id UUID PRIMARY KEY DEFAULT gen_random_uuid()`, `TIMESTAMPTZ`,
`REFERENCES users(id)`, tanpa trigger `updated_at` — update `updated_at` dari repo).
**Tanpa** `notification_send_logs`/`notification_send_log_results`.

```sql
-- notifications (blueprint)
CREATE TABLE notifications (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type           VARCHAR(50)  NOT NULL,
    title          VARCHAR(255) NOT NULL,
    body           TEXT         NOT NULL,
    payload        JSONB        NOT NULL DEFAULT '{}',
    reference_id   VARCHAR(255),
    reference_type VARCHAR(100),
    audience_type  VARCHAR(20)  NOT NULL DEFAULT 'unicast',
    audience_data  JSONB        NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_notif_audience_type ON notifications(audience_type);

-- delivery_attempts (1 baris = 1 user × 1 channel)
CREATE TABLE delivery_attempts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID        NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id)         ON DELETE CASCADE,
    channel         VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL CHECK (status IN ('pending','success','failed','retrying')),
    provider_code   VARCHAR(100),
    retry_count     INT         NOT NULL DEFAULT 0,
    next_retry_at   TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,
    attempted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_delivery_user_inapp        ON delivery_attempts(user_id, attempted_at DESC) WHERE channel = 'in_app';
CREATE INDEX idx_delivery_user_inapp_unread ON delivery_attempts(user_id) WHERE channel = 'in_app' AND read_at IS NULL;
CREATE INDEX idx_delivery_retry_at          ON delivery_attempts(next_retry_at) WHERE status = 'retrying';

-- notification_preferences
CREATE TABLE notification_preferences (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    all_enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    do_not_disturb     BOOLEAN NOT NULL DEFAULT FALSE,
    dnd_start_time     VARCHAR(5),
    dnd_end_time       VARCHAR(5),
    module_preferences JSONB   NOT NULL DEFAULT '{}',  -- fase 2
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- device_registrations — fase 2
```

`down.sql`: DROP tabel dalam urutan dependensi terbalik.

## Struktur Module

```
internal/modules/notification/
├── module.go                 -- facade; RegisterRoutes + RegisterMessageHandlers
├── contract.go               -- kosong/minimal di fase 1 (YAGNI; lihat note)
├── domain/
│   ├── notification/{constant,entity,repository}
│   ├── delivery/{constant,entity,repository}
│   ├── preference/{constant,entity,repository}
│   ├── device/{constant,entity,repository}       -- fase 2
│   └── valueobject/notification_vo.go
├── application/
│   ├── dispatcher.go          -- orchestrator send (adaptasi service/notification/dispatcher.go)
│   ├── command/
│   │   ├── send_notification.go   -- Send (unicast/multicast/broadcast) via dispatcher
│   │   ├── mark_read.go
│   │   ├── mark_all_read.go
│   │   └── update_preference.go   -- global + DND (fase 2: per-modul)
│   ├── query/
│   │   ├── list_inbox.go
│   │   ├── unread_count.go
│   │   └── get_preference.go
│   ├── dto/                   -- NotificationItem, UnreadCountResponse, PreferenceRequest/Response
│   ├── errors.go              -- mapping domain -> ERR_*
│   └── ports/
│       ├── identity_reader.go  -- GetUserSummary (fase 2, email/sms)
│       ├── transactor.go
│       └── channel_sender.go   -- EmailSender/SMSSender/PushSender (fase 2)
├── infrastructure/
│   ├── persistence/          -- postgres_*_repo + transactor + helpers
│   ├── identitygateway/      -- identity.Contract -> ports.IdentityReader (fase 2)
│   └── external/             -- smtp_emailer / fonnte_sms / fcm (fase 2)
└── interfaces/
    ├── http/
    │   ├── handler.go
    │   └── router.go
    └── mq/                    -- event consumer (pola akademik/interfaces/mq)
        ├── routing.go         -- RoutingLoginSucceeded = "identity.user.login_succeeded" + Bindings
        ├── payload.go         -- LoginSucceededPayload{UserID}
        ├── handler.go         -- decode -> dispatch
        └── register.go        -- RegisterHandlers(registry, deps)
```

Pemetaan file blueprint → target (yang relevan fase 1):

| k-forum-api | SIPON target |
|---|---|
| `domain/notification/entity/notification.go` | `domain/notification/entity/notification.go` |
| `domain/notification/entity/delivery_attempt.go` | `domain/delivery/entity/delivery_attempt.go` |
| `domain/notification/entity/notification_module_preference.go` | `domain/preference/entity/notification_preference.go` (disederhanakan) |
| `domain/notification/entity/notification_send_log.go` | **dibuang** (skema tak berguna) |
| `domain/notification/constant/notification_constant.go` | `domain/notification/constant/notification_constant.go` |
| `domain/notification/valueobject/notification_vo.go` | `domain/valueobject/notification_vo.go` |
| `domain/notification/repository/interfaces.go` | `domain/*/repository/interfaces.go` (dipecah per aggregate) |
| `domain/notification/service/module_preference_gate.go` | `domain/preference/service/module_preference_gate.go` (fase 2) |
| `app/service/notification/dispatcher.go` | `application/dispatcher.go` |
| `app/usecase/notification/{list,mark_read,get_unread_count,mark_all_read}.go` | `application/{query,command}/*` |
| `app/usecase/notificationpreference/*.go` | `application/command/update_preference.go` + `query/get_preference.go` |
| `app/usecase/notificationtester/*` | **dibuang** (test kirim tanpa riwayat) |
| `app/dto/notification_dto.go` | `application/dto/notification_dto.go` |
| `app/port/notification_query_model.go` | `domain/delivery/repository/interfaces.go` (read model) |
| `app/port/notification_hub.go` | (fase 3) `application/ports/realtime.go` |
| `interfaces/http/handler/.../notification_handler.go` | `interfaces/http/handler.go` |
| `interfaces/mq/handler/notification_handler.go` | `interfaces/mq/*` (pola akademik) |

## Event (Routing Key & Payload)

Mengikuti konvensi `<module>.<resource>.<action>` (lih. `akademik/interfaces/mq/routing.go`):

```go
// interfaces/mq/routing.go
const (
    RoutingLoginSucceeded = "identity.user.login_succeeded"
    QueueNotification     = "sipon.worker.notification"
)

var Bindings = []messaging.Binding{
    {Queue: QueueNotification, RoutingKey: RoutingLoginSucceeded},
}
```

```go
// interfaces/mq/payload.go
type LoginSucceededPayload struct {
    UserID string `json:"user_id"`
}
func (p LoginSucceededPayload) Validate() error // user_id wajib
```

`interfaces/mq/register.go` mendaftarkan handler ke `messaging.Contract` dan
mengembalikan `Bindings`. Handler decode → `Validate` → panggil dispatcher →
bungkus error sebagai `messaging.NewFatalError`/`NewRetryableError` (pola
`akademik`). Setelah login-succeeded jalan, module lain tinggal menambah event
serupa: `keuangan.payment.verified`, `psb.review.approved`, dst.

## Alur Pengiriman (Dispatcher)

Adaptasi `dispatcher.go` k-forum-api, disederhanakan untuk fase 1 (in-app only):

1. Handler MQ → `application/dispatcher.go#Dispatch(tmpl, target)`.
2. `unicast`: buat blueprint `Notification`, lalu dalam satu transaksi simpan
   blueprint + `DeliveryAttempt` `in_app` (status `success` — in-app tidak punya
   kegagalan I/O).
3. `multicast`/`broadcast`: satu blueprint + fanout `DeliveryAttempt` per penerima,
   dibatch (batch 50, pola k-forum-api). Broadcast admin dipicu HTTP handler
   notification sendiri (bukan event).
4. Fase 2: setelah in_app, kirim channel lain (`email`/`sms`/`push`) dengan
   `DeliveryAttempt` masing-masing + status success/failed.

## Preferensi & Filter

Fase 1:
- `AllEnabled == false` → skip semua channel (cek sebelum bikin DeliveryAttempt).
- DND (`IsWithinDND`) → relevan saat ada channel realtime/push; simpan struktur
  sekarang, terapkan penuh di fase 2.
- `Bypass == true` → lewati semua filter (notifikasi operasional penting).

Fase 2: `ModulePreferenceGate` (adaptasi `module_preference_gate.go`) dengan
`eventKey = "{module}.{event}"` dan map modul SIPON (`keuangan`, `psb`, `akademik`,
`kesantrian`, `announcement`), `IsAllowedBulk` untuk fanout broadcast.

## API Endpoints (JWT, `principalLoad`)

Self-service (`/api/v1/web/notifications`):
- `GET /inbox` — list notifikasi in_app (filter `unread_only`, pagination)
- `GET /unread-count` — badge
- `POST /:id/read` — tandai satu dibaca (ownership check)
- `POST /read-all` — tandai semua dibaca
- `GET /preferences` — ambil preferensi (auto-create default bila belum ada)
- `PUT /preferences` — update global + DND

Admin (`manage_notification` / broadcast pengumuman):
- `POST /api/v1/web/notification/admin/broadcast` — kirim broadcast (tanpa riwayat)

Permission baru `manage_notification` ditambahkan ke `identity` (pola
`manage_feedback`): masuk `AllPermissionDefinitions`, default di system role
`usergod`/`superadmin`/`admin`, assignable ke custom role.

## Business Rules

1. User login bisa lihat inbox & unread count miliknya saja.
2. Mark read hanya untuk notifikasi milik sendiri (ownership di SQL → 404 bila bukan).
3. Preferensi auto-create saat pertama diakses (default semua enabled).
4. DND validasi format `HH:MM`; saat `do_not_disturb=true` kedua time wajib ada.
5. `Bypass` melewati `AllEnabled`/DND/preferensi modul (operasional penting).
6. Broadcast admin tidak menyimpan riwayat (test kirim = test saja).

## Fase Implementasi

1. **Fase 1 (MVP)** — migration (notification/delivery/preference), domain +
   persistence, dispatcher in-app, HTTP inbox/preference/read, broadcast admin
   (tanpa send log), **event login-succeeded** (identity publish + notification
   consume), wiring `cmd/api/main.go` (outbox ke identity) & `cmd/worker/main.go`
   (RegisterMessageHandlers), permission `manage_notification`.
2. **Fase 2** — channel email/sms (reuse infra identity), device_registrations +
   push, preferensi per-modul (`ModulePreferenceGate`).
3. **Fase 3** — realtime (SSE/WS), retry `delivery_attempts` bila channel push
   membutuhkannya.

## Verifikasi

1. `go build ./...`, `go vet ./...`, `go test ./internal/modules/notification/...` lolos.
2. Migration `up`/`down` bersih.
3. Smoke test E2E: login berhasil → outbox berisi `identity.user.login_succeeded`
   → worker relay/consume → notif muncul di inbox user → unread-count bertambah →
   mark read → count berkurang → read-all → kosong; admin broadcast → semua user
   dapat notif; preferensi off → notif tidak terkirim (kecuali bypass).
