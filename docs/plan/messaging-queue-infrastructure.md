# Penyempurnaan Infrastruktur Messaging Queue

## Status

Proposed. Dokumen ini adalah rencana implementasi, bukan perubahan kode yang sudah
dikerjakan.

## Context

`sipon-be` saat ini sudah memiliki fondasi scheduler yang cukup untuk menjalankan
background job, tetapi belum memiliki message queue yang reliable:

- `internal/shared/scheduler/application/worker.go` mengambil job dari
  `scheduled_jobs`, lalu langsung men-dispatch handler di proses worker.
- `internal/shared/scheduler/infrastructure/persistence/postgres_scheduled_job_repository.go`
  sudah memakai `FOR UPDATE SKIP LOCKED`, tetapi belum memiliki lease/timeout untuk
  memulihkan job yang tertinggal dalam status `PROCESSING`.
- `internal/shared/eventbus/eventbus.go` masih berupa stub dan belum dipakai module.
- `go.mod` belum memiliki dependency AMQP/RabbitMQ.
- `docker-compose.dev.yml` hanya menyediakan PostgreSQL, Redis, MinIO, API, dan
  scheduler worker.
- Sebagian besar `RegisterSchedulerHandlers` masih no-op; handler nyata yang terlihat
  saat ini terutama berada di module akademik.

Referensi utama:

- `k-forum-api/internal/infrastructure/external/rabbitmq/client.go`
- `k-forum-api/internal/interfaces/mq/relay/outbox_relay.go`
- `k-forum-api/internal/interfaces/mq/relay/scheduler_relay.go`
- `k-forum-api/internal/interfaces/mq/middleware/job_lifecycle.go`
- `k-forum-api/internal/app/service/outbox/outbox_entry.go`
- `k-forum-api/internal/app/service/job/job.go`
- `k-forum-api/internal/interfaces/mq/message.go`

## Goals

- Memisahkan scheduling, publishing, dan eksekusi handler.
- Menjamin perubahan data bisnis dan pencatatan event terjadi atomik melalui
  transactional outbox.
- Menjamin message tidak hilang ketika RabbitMQ atau worker sedang down.
- Menyediakan retry yang bounded, memiliki backoff, dan berakhir di DLQ bila tidak
  dapat diproses.
- Menjadikan handler idempotent terhadap redelivery dan crash setelah side effect.
- Mendukung lebih dari satu worker tanpa double claim.
- Menyediakan correlation ID, message ID, status, attempt count, dan error terakhir
  untuk troubleshooting.
- Mempertahankan batas module: module hanya mengenal port/contract messaging dan
  routing key, bukan client RabbitMQ secara langsung.

## Non-goals

- Tidak mengganti PostgreSQL sebagai persistence utama aplikasi.
- Tidak memakai Redis sebagai durable queue.
- Tidak memindahkan seluruh business logic ke consumer layer.
- Tidak menambahkan WebSocket atau push notification pada fase awal. Adapter tersebut
  dapat ditambahkan kemudian jika SIPON memiliki consumer lintas-proses yang nyata.
- Tidak mengubah semua usecase menjadi asynchronous sekaligus. Migrasi dimulai dari
  scheduled job dan satu use case pilot.

## Referensi Yang Diambil

| Pola dari `k-forum-api` | Adaptasi untuk `sipon-be` |
| --- | --- |
| Durable topic exchange dan queue per consumer role | Dipakai sebagai transport utama event asynchronous |
| `event_outbox` sebagai buffer publish | Ditambahkan dan ditulis dalam transaksi yang sama dengan perubahan bisnis |
| Envelope `job_id`, `type`, `version`, `occurred_at`, `payload`, `correlation_id` | Dipakai sebagai kontrak transport versi pertama |
| Registry dan middleware lifecycle | Dipakai untuk memisahkan routing, status job, dan handler module |
| `scheduled_jobs` sebagai sumber jadwal | Dipertahankan, tetapi relay hanya membuat outbox message; tidak mengeksekusi handler langsung |
| Job record terpisah dari outbox | Ditambahkan sebagai `message_jobs` untuk idempotency dan audit eksekusi |
| DLQ per queue | Dipakai untuk pesan fatal atau retry yang sudah habis |

Beberapa detail referensi tidak disalin apa adanya:

- Publisher harus memakai broker publisher confirm sebelum outbox ditandai
  `PUBLISHED`.
- Channel AMQP tidak boleh dipakai bersamaan tanpa serialisasi karena `Publish` dan
  consumer registration dapat mengalami race.
- `Nack(false, true)` langsung ke queue utama dapat membuat hot-loop dan tidak memberi
  delay retry. Retry SIPON harus melewati retry queue dengan TTL atau mekanisme delay
  yang setara.
- `x-death` hanya dijadikan metadata pendukung, bukan satu-satunya sumber kebenaran
  attempt count. Attempt count juga dicatat di `message_jobs`.

## Target Architecture

```text
HTTP API / Module Usecase
        |
        | DB Transaction
        |   +-- business write
        |   +-- event_outbox INSERT
        v
   PostgreSQL
        |
        | Outbox Relay
        | claim PENDING events
        v
RabbitMQ Topic Exchange
        |
        +--> sipon.worker.scheduler
        |       |
        |       +--> Consumer
        |       +--> Retry Queue (TTL)
        |       +--> DLQ
        |
        +--> sipon.worker.notifications
        |       |
        |       +--> Consumer
        |       +--> Retry Queue (TTL)
        |       +--> DLQ
        |
        +--> sipon.worker.<module>
                |
                +--> Consumer
                +--> Retry Queue (TTL)
                +--> DLQ


Worker Process
  |
  +-- Scheduler Dispatcher
  |       scheduled_jobs
  |              |
  |              v
  |       event_outbox
  |
  +-- Outbox Relay
  |       event_outbox
  |              |
  |              v
  |       RabbitMQ
  |
  +-- Message Consumer
          RabbitMQ
              |
              v
        message_jobs
              |
              v
        Module Handler
```

Prinsip penting:

- `scheduled_jobs` menjawab kapan pekerjaan harus dipicu.
- `event_outbox` dan business write selalu berada dalam DB transaction yang sama.
- `Scheduler Dispatcher` hanya mengubah scheduled job dan membuat outbox event; ia
  tidak mengirim ke RabbitMQ dan tidak menjalankan module handler.
- `Outbox Relay` bertanggung jawab atas `event_outbox -> RabbitMQ`.
- Retry queue dengan TTL dan DLQ adalah bagian dari RabbitMQ consumer topology, bukan
  bagian dari business logic worker.
- `message_jobs` adalah durable inbox/message record yang dicatat sebelum handler
  diproses, sehingga redelivery dapat di-deduplicate.
- Handler hanya dianggap sukses setelah side effect dan status `SUCCEEDED` berhasil
  ditulis.

## Existing Scheduler Execution Before MQ

### Alur aktual pada module akademik

Saat ini eksekusi job akademik berjalan seperti berikut:

1. `OpenSessionUseCase.Execute` membuka session dan menyimpan perubahan session.
2. Usecase tersebut memanggil `ScheduleSessionJobsUseCase.Execute`.
3. Usecase scheduler membuat dua row di `scheduled_jobs`:
   `akademik.fingerprint_sync` sebagai recurring job setiap menit dan
   `akademik.session_auto_close` sebagai one-off job pada `ends_at`.
4. `cmd/worker/main.go` membuat satu `scheduler.Registry` lalu memanggil
   `RegisterSchedulerHandlers` dari semua module.
5. `scheduler.Worker` melakukan `FindDueAndClaim`, lalu memanggil
   `registry.Dispatch(ctx, job.Type, job.Payload)` secara langsung.
6. `Module.handleFingerprintSync` atau `Module.handleSessionAutoClose` melakukan
   parsing JSON, memanggil usecase akademik, lalu mengembalikan error retryable atau
   fatal.
7. Setelah handler selesai, worker mengubah row scheduled job menjadi completed,
   active untuk recurring job, atau failed.
8. Handler auto-close juga mencari recurring fingerprint job berdasarkan
   `reference_id`, lalu melakukan pause secara langsung.

### Masalah dari pola aktual

- Handler, parsing payload, routing key, dan klasifikasi error bercampur di
  `internal/modules/akademik/module.go`.
- `RegisterSchedulerHandlers` berbeda nama dan mekanisme dari model consumer MQ yang
  direncanakan.
- `scheduler.Registry` hanya mengenal `jobType` dan `json.RawMessage`, sehingga belum
  membawa message ID, version, occurred time, atau correlation ID.
- Worker scheduler dan consumer MQ nantinya berpotensi mendaftarkan handler yang
  sama dua kali dengan kontrak berbeda.
- `ScheduleSessionJobsUseCase` menyimpan recurring job lalu one-off job secara
  terpisah; kegagalan save kedua dapat meninggalkan schedule parsial.
- `OpenSessionUseCase` hanya me-log kegagalan scheduling dan tetap mengembalikan
  session sukses, sehingga caller tidak mendapat sinyal bahwa background lifecycle
  gagal dibuat.
- `handleSessionAutoClose` mengubah scheduler repository dari handler adapter. Pada
  desain baru, operasi pause harus menjadi tanggung jawab application usecase atau
  application port, bukan layer transport.

## Uniform Module `interfaces/mq` Layer

Setiap module yang memiliki asynchronous handler wajib menyediakan adapter MQ dengan
struktur seragam:

```text
internal/modules/<module>/interfaces/mq/
├── routing.go       # routing key dan binding queue module
├── payload.go       # DTO payload message dan validasi dasar
├── handler.go       # handler adapter: decode -> call usecase -> classify error
└── register.go      # RegisterHandlers(registry) dan daftar binding
```

Contoh untuk akademik:

```text
internal/modules/akademik/interfaces/mq/
├── routing.go
├── payload.go
├── handler.go
└── register.go
```

### Ownership dan dependency rule

- `interfaces/mq` adalah inbound adapter module, sama levelnya dengan
  `interfaces/http`.
- Routing key dan queue binding dimiliki `interfaces/mq`, bukan `cmd/worker` dan
  bukan file `application/command/job_types.go`.
- Payload DTO message dimiliki `interfaces/mq`; domain entity tidak boleh menjadi
  payload transport.
- Handler MQ hanya menerjemahkan transport ke application usecase.
- Business rule, transaksi, pause/resume schedule, dan idempotency bisnis berada di
  application/domain layer.
- `interfaces/mq` boleh bergantung pada shared messaging port dan application port
  module, tetapi tidak boleh membuat connection RabbitMQ atau mengakses channel AMQP.
- `cmd/worker` hanya melakukan composition: construct module, register handlers, dan
  start relay/consumer. Ia tidak boleh mengetahui routing key satu per satu.

### Contract registration yang seragam

Shared messaging package menyediakan binding dan registry. Setiap module meng-expose
satu method publik pada `Module`:

```go
func (m *Module) RegisterMessageHandlers(
    registry *messaging.Registry,
) ([]messaging.Binding, error)
```

Implementasinya mendelegasikan ke adapter module:

```go
func (m *Module) RegisterMessageHandlers(
    registry *messaging.Registry,
) ([]messaging.Binding, error) {
    return akademikmq.RegisterHandlers(registry, akademikmq.Dependencies{
        FingerprintSync: m.syncFingerprintUC,
        SessionAutoClose: m.autoCloseSessionUC,
    })
}
```

`RegisterHandlers` harus:

- gagal jika dua handler mendaftarkan routing key yang sama;
- mendaftarkan handler dan mengembalikan binding queue dalam satu operasi;
- tidak melakukan network call atau membaca database saat startup registration;
- menggunakan shared `messaging.RetryableError` dan `messaging.FatalError`;
- dapat dipakai oleh direct-dispatch bridge maupun RabbitMQ consumer.

Contoh routing dan binding akademik:

```go
const (
    RoutingFingerprintSync = "akademik.fingerprint.sync"
    RoutingSessionAutoClose = "akademik.session.auto_close"
    QueueScheduler = "sipon.worker.scheduler"
)

var Bindings = []messaging.Binding{
    {Queue: QueueScheduler, RoutingKey: RoutingFingerprintSync},
    {Queue: QueueScheduler, RoutingKey: RoutingSessionAutoClose},
}
```

Semua module memakai format routing key yang sama: `<module>.<resource>.<action>`.
Queue ditentukan berdasarkan consumer role, bukan satu queue baru untuk setiap
handler.

Karena `scheduled_jobs.type` sudah dipersist, perubahan dari routing key lama harus
ditangani sebagai migration data, bukan sekadar mengganti konstanta:

- Backfill `akademik.fingerprint_sync` menjadi `akademik.fingerprint.sync`.
- Backfill `akademik.session_auto_close` menjadi `akademik.session.auto_close`.
- Selama rollout, registry boleh mendaftarkan alias legacy sebagai compatibility
  window yang jelas batas penghapusannya.
- Setelah seluruh row lama ter-backfill dan tidak ada deployment lama yang berjalan,
  alias legacy dihapus.

### Akademik handler adapter

Pindahkan `handleFingerprintSync` dan `handleSessionAutoClose` dari `module.go` ke
`internal/modules/akademik/interfaces/mq/handler.go`:

- `FingerprintSyncPayload{SessionID string}` dan
  `SessionAutoClosePayload{SessionID string}` menjadi DTO typed.
- Decode dan validasi `session_id` dilakukan di adapter; payload invalid menjadi
  `FatalError`.
- `FingerprintSyncHandler` memanggil application usecase sync dan membungkus error
  transient sebagai `RetryableError`.
- `SessionAutoCloseHandler` hanya memanggil `AutoCloseSessionUseCase`.
- Pause recurring fingerprint job dipindahkan ke `AutoCloseSessionUseCase` atau port
  application scheduler yang dipanggil usecase tersebut.
- Handler menerima `Message` lengkap sehingga log/usecase dapat memakai message ID dan
  correlation ID.

Tambahkan `AutoCloseSessionUseCase` bila logic pause belum dapat ditempatkan dengan
jelas di `CompleteSessionUseCase`. Tujuannya bukan membuat usecase baru tanpa alasan,
melainkan mengeluarkan business workflow dari transport adapter dan memastikan
complete session + pause schedule memiliki boundary yang dapat dites.

### Routing key untuk scheduled job

`scheduled_jobs.type` harus berisi routing key canonical yang sama dengan envelope
`Message.Type`. Dengan begitu scheduler tidak memiliki kamus routing kedua.

`JobTypeFingerprintSync` dan `JobTypeSessionAutoClose` yang sekarang berada di
`application/command/job_types.go` dipindahkan atau dihapus setelah constructor
scheduler menerima konfigurasi route dari module composition:

```go
scheduleJobsUC := command.NewScheduleSessionJobsUseCase(
    scheduledJobRepo,
    cronParser,
    timeutil.Loc(),
    command.ScheduledJobTypes{
        FingerprintSync: akademikmq.RoutingFingerprintSync,
        SessionAutoClose: akademikmq.RoutingSessionAutoClose,
    },
)
```

`ScheduleSessionJobsUseCase` hanya menyimpan string routing key yang diinjeksi; ia
tidak meng-import `interfaces/mq`. Ini menjaga arah dependency application -> port,
sementara composition root yang menyatukan command dan adapter.

### Compatibility bridge sebelum RabbitMQ aktif

Refactor tidak boleh membuat handler akademik ditulis dua kali. Urutannya:

1. Tambahkan shared `messaging.Registry` dan pindahkan handler akademik ke
   `interfaces/mq`.
2. Ubah scheduler worker sementara agar dispatch ke `messaging.Message`, bukan
   `scheduler.Registry` lama.
3. Scheduler direct-dispatch bridge membuat envelope dari row `scheduled_jobs`, lalu
   memakai registry yang sama dengan calon consumer MQ.
4. Setelah outbox dan RabbitMQ consumer aktif, hapus direct-dispatch bridge; handler
   dan registration module tidak berubah lagi.
5. Hapus `RegisterSchedulerHandlers`, `scheduler.Registry`, serta error wrapper lama
   setelah seluruh module selesai migrasi.

Dengan pola ini, peralihan transport adalah perubahan di worker wiring, bukan
perubahan ulang pada setiap handler module.

## Message Contract

Tambahkan envelope transport yang stabil di shared messaging package:

```go
type Message struct {
    ID            uuid.UUID       `json:"id"`
    Type          string          `json:"type"`
    Version       int             `json:"version"`
    OccurredAt    time.Time       `json:"occurred_at"`
    Payload       json.RawMessage `json:"payload"`
    CorrelationID string          `json:"correlation_id"`
    CausationID   *uuid.UUID      `json:"causation_id,omitempty"`
}
```

Ketentuan kontrak:

- `ID` adalah idempotency key dan tidak berubah ketika message di-retry.
- `Type` memakai routing key lowercase dengan format `<module>.<action>`, misalnya
  `akademik.session.auto_close`.
- `Version` dimulai dari `1`; perubahan breaking harus menaikkan version atau membuat
  routing key baru.
- `Payload` hanya berisi DTO event, bukan entity domain atau object repository.
- `CorrelationID` berasal dari request ID; job dari scheduler membuat UUID baru.
- `CausationID` menghubungkan message turunan dengan message pemicu.

## Data Model

### 1. `event_outbox`

Buat migration baru, tanpa mengubah migration historis. Minimal kolom:

```sql
CREATE TABLE event_outbox (
    id              UUID PRIMARY KEY,
    routing_key     VARCHAR(150) NOT NULL,
    payload         JSONB NOT NULL,
    version         INT NOT NULL DEFAULT 1,
    correlation_id  VARCHAR(64),
    causation_id    UUID,
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempt_count   INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at       TIMESTAMPTZ,
    published_at    TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Status yang diizinkan: `PENDING`, `PUBLISHING`, `PUBLISHED`, `FAILED`.

Tambahkan index partial untuk `(next_attempt_at, created_at)` pada status
`PENDING` dan `FAILED`. Relay harus melakukan claim dengan `FOR UPDATE SKIP LOCKED`.
Row `PUBLISHING` yang melewati lease timeout dikembalikan ke `PENDING`.

Transactional boundary yang wajib dijaga adalah:

```text
BEGIN
  business write
  INSERT event_outbox
COMMIT
```

Keduanya harus commit atau rollback bersama. Publish ke RabbitMQ dilakukan setelah
commit oleh `Outbox Relay/Publisher`, bukan di dalam request transaction.

### 2. `message_jobs`

Tabel ini berbeda dari `scheduled_jobs`. `scheduled_jobs` adalah definisi waktu,
sedangkan `message_jobs` adalah durable inbox dan lifecycle setiap delivery.
Consumer wajib membuat atau memastikan row inbox ini durable sebelum handler
menjalankan side effect.

```sql
CREATE TABLE message_jobs (
    id              UUID PRIMARY KEY,
    routing_key     VARCHAR(150) NOT NULL,
    payload         JSONB NOT NULL,
    version         INT NOT NULL,
    correlation_id  VARCHAR(64),
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    attempt_count   INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 5,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    running_at      TIMESTAMPTZ,
    succeeded_at    TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    locked_until    TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Status yang diizinkan: `PENDING`, `RUNNING`, `RETRY_WAIT`, `SUCCEEDED`, `FAILED`.

State transition normal:

```text
PENDING -> RUNNING -> SUCCEEDED
RUNNING -> RETRY_WAIT -> RUNNING
RUNNING -> FAILED
```

`INSERT ... ON CONFLICT (id) DO NOTHING` wajib dipakai ketika message pertama kali
diterima dan transaction insert harus commit sebelum handler dijalankan. Jika row
sudah `SUCCEEDED`, consumer harus ack tanpa menjalankan handler lagi. Jika row
`FAILED`, consumer tidak boleh mengulang side effect tanpa operasi replay yang
eksplisit.

### 3. Penyempurnaan `scheduled_jobs`

Tambahkan lease dan recovery untuk job yang diklaim tetapi worker crash:

- `lease_until TIMESTAMPTZ`.
- `attempt_count` bila `retry_count` yang ada belum cukup jelas secara semantik.
- index untuk job due dan job `PROCESSING` yang lease-nya expired.

`FindDueAndClaim` harus juga merebut kembali job `PROCESSING` dengan `lease_until <=
NOW()`. Retry tidak boleh memakai `next_run_at` lama karena akan menimbulkan polling
tanpa backoff.

## Component Design

### Shared messaging ports

Buat package shared, misalnya `internal/shared/messaging`, yang berisi:

- `Message` dan error classification.
- `Publisher` port dengan method `Publish(ctx, Message) error`.
- `Binding` dan contract untuk daftar queue/routing key yang dimiliki module.
- `HandlerFunc` dan `Registry`.
- `RetryableError` dan `FatalError`.
- retry policy dan fungsi exponential backoff dengan jitter.
- konstanta exchange, status, dan metadata header yang tidak bergantung pada AMQP.

RabbitMQ adapter berada di `internal/shared/messaging/infrastructure/rabbitmq`.
Module tidak boleh meng-import package RabbitMQ secara langsung.

### RabbitMQ consumer topology

Retry queue dan DLQ adalah bagian dari topology RabbitMQ yang mendukung setiap
consumer, bukan tanggung jawab business handler atau scheduler worker. Gunakan
topology durable berikut:

- Exchange utama: `sipon.events`, type `topic`, durable.
- Exchange dead-letter: `sipon.events.dlx`, type `direct`, durable.
- Queue utama dibuat per consumer role, bukan satu queue global.
- Setiap queue utama memiliki DLX dan routing key DLQ sendiri.
- Retry queue menggunakan TTL bertingkat, misalnya `1m`, `5m`, dan `30m`, lalu
  dead-letter kembali ke exchange utama.
- Semua published message memakai `ContentType=application/json` dan
  `DeliveryMode=Persistent`.
- Consumer hanya menjalankan proses consume dan acknowledgement; deklarasi main
  queue, retry queue, TTL, DLX, dan DLQ dilakukan oleh adapter topology RabbitMQ.

Nama queue harus stabil dan didefinisikan sebagai konstanta. Queue tidak boleh
dibuat berdasarkan request ID, entity ID, atau nilai payload.

### Publisher

Implementasi publisher harus:

- Membuka connection dan channel khusus publisher.
- Mengaktifkan publisher confirm dan menunggu confirmation broker.
- Memiliki timeout publish yang configurable.
- Menangani reconnect ketika connection/channel ditutup.
- Menserialisasi publish pada satu channel atau memakai channel pool yang aman.
- Tidak menandai outbox `PUBLISHED` sebelum publish confirm berhasil.
- Meneruskan `message_id`, `correlation_id`, `type`, dan `version` sebagai header
  untuk tracing dan diagnosis.

### Outbox Relay/Publisher

Implementasikan `OutboxRelay` di worker:

1. Claim batch outbox berstatus `PENDING` atau retryable `FAILED` yang
   `next_attempt_at <= NOW()`.
2. Set status `PUBLISHING` dan lease dalam transaksi singkat.
3. Publish envelope ke RabbitMQ dengan confirm.
4. Jika confirm sukses, set `PUBLISHED` dan `published_at`.
5. Jika gagal, increment `attempt_count`, simpan `last_error`, hitung
   `next_attempt_at`, lalu set kembali `FAILED`.
6. Jika jumlah attempt melewati batas, jangan hapus event. Tandai `FAILED` dan
   expose sebagai operational alert/replay candidate.

Outbox Relay/Publisher harus idempotent: crash setelah publish tetapi sebelum update database
dapat menyebabkan duplicate delivery, sehingga consumer wajib idempotent.

### Scheduler Dispatcher

Refactor `internal/shared/scheduler/application/worker.go` sehingga fungsi scheduling
menjadi `Scheduler Dispatcher`, bukan relay:

- `Scheduler Dispatcher`: claim due `scheduled_jobs`, lalu membuat satu
  `event_outbox`.
- `Outbox Relay/Publisher`: claim dan publish `event_outbox` ke RabbitMQ.
- `Message Consumer`: consume RabbitMQ, persist inbox `message_jobs`, lalu dispatch
  ke handler.

Untuk setiap scheduled job, insert outbox dan update state scheduled job dilakukan
dalam transaksi database yang sama:

- One-off: `PROCESSING -> COMPLETED` setelah outbox berhasil dibuat.
- Recurring: `PROCESSING -> ACTIVE` dan hitung `next_run_at` berikutnya.
- Jika transaksi rollback, job tetap dapat di-claim ulang.
- Scheduler Dispatcher tidak pernah memanggil handler module secara langsung.

### Consumer lifecycle

Consumer menggunakan manual acknowledgement dan prefetch configurable. Alur
durable inbox adalah `RabbitMQ -> message_jobs -> Module Handler`:

1. Parse envelope dan validasi `ID`, `Type`, `Version`, serta payload.
2. Dalam DB transaction, insert durable inbox `message_jobs` dengan status `PENDING`
   menggunakan `ON CONFLICT (id) DO NOTHING`, lalu commit sebelum processing.
3. Jika row sudah `SUCCEEDED`, ack dan selesai tanpa memanggil handler.
4. Jika row `FAILED`, ack dan selesai; replay harus melalui command operasional yang
   membuat attempt/replay baru.
5. Claim row `PENDING` dengan lease, ubah menjadi `RUNNING`, lalu commit sebelum
   dispatch. Jika lease `RUNNING` masih aktif, jangan jalankan duplicate paralel.
6. Dispatch message ke registry dan module handler.
7. Jika handler sukses, ubah state menjadi `SUCCEEDED` secara durable, kemudian
   `Ack` message.
8. Jika handler mengembalikan `FatalError`, ubah state menjadi `FAILED`, lalu ack
   atau nack ke DLQ sesuai konfigurasi topology. State database harus sudah tercatat.
9. Jika error retryable dan attempt masih tersedia, publish ulang ke retry queue
   dengan TTL. Setelah publish retry terkonfirmasi dan state menjadi `RETRY_WAIT`,
   ack message asli.
10. Jika retry habis, ubah state menjadi `FAILED` dan route message ke DLQ.
11. Panic handler harus di-recover per message dan diperlakukan sebagai retryable
    sampai batas attempt tercapai.

## Phased Implementation Plan

Roadmap dibagi berdasarkan satu perubahan arsitektural per fase. Setiap fase memiliki
exit criteria sehingga fase berikutnya tidak dimulai ketika kontrak atau state
sebelumnya belum stabil.

### Phase 0: Baseline and decisions

Tujuan: mengunci scope, routing, dan perilaku existing sebelum refactor.

Pekerjaan:

- Dokumentasikan alur akademik existing: `OpenSessionUseCase` membuat recurring
  fingerprint sync dan one-off auto-close, lalu worker direct-dispatch.
- Inventarisasi semua `RegisterSchedulerHandlers`, job type, payload, dan handler
  nyata di setiap module.
- Tetapkan consumer role awal: `sipon.worker.scheduler` untuk scheduled job.
- Tetapkan format routing key `<module>.<resource>.<action>`.
- Tetapkan max attempts, retry delay, batch size, prefetch, lease timeout, retention,
  dan timeout handler.
- Tetapkan migration/backfill untuk routing key legacy yang sudah tersimpan di
  `scheduled_jobs`.

Deliverables:

- Daftar routing key dan payload awal.
- Diagram final dan keputusan arsitektur yang sudah disepakati.
- Test baseline untuk status transition scheduled job.

Exit criteria: tidak ada routing key atau handler scheduled yang belum memiliki owner
module dan payload contract.

### Phase 1: Shared messaging contract

Tujuan: membuat kontrak transport dan registry tanpa bergantung pada RabbitMQ.

Pekerjaan:

- Tambahkan `Message`, `Binding`, `HandlerFunc`, dan `Registry` di
  `internal/shared/messaging`.
- Tambahkan `RetryableError` dan `FatalError` yang dipakai semua module.
- Tambahkan validation untuk `ID`, `Type`, `Version`, `OccurredAt`, dan payload.
- Pastikan registry menolak duplicate routing key.
- Tambahkan unit test registry, error classification, envelope, dan binding.

Deliverables:

- Shared message contract.
- Registry yang dapat dipakai direct-dispatch maupun RabbitMQ consumer.

Exit criteria: package messaging lulus `go test` dan tidak meng-import package AMQP.

### Phase 2: Module `interfaces/mq` extraction

Tujuan: menyeragamkan registration dan routing sebelum transport RabbitMQ masuk.

Pekerjaan:

- Tambahkan struktur `routing.go`, `payload.go`, `handler.go`, dan `register.go` pada
  module yang memiliki asynchronous handler.
- Tambahkan facade `Module.RegisterMessageHandlers` yang mengembalikan binding.
- Refactor akademik sebagai module pertama.
- Pindahkan `JobTypeFingerprintSync` dan `JobTypeSessionAutoClose` dari
  `application/command/job_types.go` ke routing adapter melalui route configuration
  injection.
- Pindahkan handler fingerprint sync dan session auto-close dari `module.go` ke
  `internal/modules/akademik/interfaces/mq/`.
- Pindahkan typed payload, parsing, dan error classification ke adapter MQ.
- Pindahkan pause recurring job dari transport handler ke application usecase atau
  application port.
- Pertahankan alias routing key legacy hanya selama migration data.

Deliverables:

- `internal/modules/akademik/interfaces/mq/` yang lengkap.
- Contract registration yang sama untuk setiap module.
- Tidak ada routing key akademik yang disimpan manual di `cmd/worker`.

Exit criteria: akademik dapat didaftarkan ke shared registry tanpa
`RegisterSchedulerHandlers` dan tanpa koneksi RabbitMQ.

### Phase 3: Persistence and transaction foundation

Tujuan: menyiapkan state persistence sebelum scheduler dan consumer dipisahkan.

Pekerjaan:

- Tambahkan migration `event_outbox` dan down migration.
- Tambahkan migration `message_jobs` dan down migration.
- Tambahkan lease/recovery column pada `scheduled_jobs`.
- Tambahkan transactor/exec abstraction agar business repository dan outbox memakai
  DB transaction yang sama.
- Tambahkan repository port dan PostgreSQL adapter untuk outbox.
- Tambahkan repository port dan PostgreSQL adapter untuk durable inbox
  `message_jobs`.
- Tambahkan index claim, status, due time, lease timeout, dan retention.

Deliverables:

- Schema `event_outbox`, `message_jobs`, dan lease scheduler.
- Repository tests untuk claim, duplicate insert, lease recovery, dan state transition.

Exit criteria: business write + `event_outbox INSERT` dapat dilakukan dalam satu
transaction yang commit/rollback bersama.

### Phase 4: Scheduler Dispatcher

Tujuan: memisahkan scheduling dari eksekusi handler.

Pekerjaan:

- Ubah worker scheduler menjadi `Scheduler Dispatcher`.
- Claim due `scheduled_jobs` dengan `FOR UPDATE SKIP LOCKED` dan lease.
- Dalam satu DB transaction, insert event ke `event_outbox` dan update scheduled job.
- One-off menjadi `COMPLETED`; recurring kembali `ACTIVE` dengan `next_run_at` baru.
- Tambahkan recovery untuk job `PROCESSING` yang lease-nya expired.
- Tambahkan direct-dispatch compatibility bridge yang membuat `messaging.Message` dari
  scheduled job dan memakai registry module yang sama.
- Scheduler Dispatcher tidak boleh memanggil business handler secara langsung setelah
  mode outbox aktif.

Deliverables:

- `Scheduler Dispatcher` yang hanya menghasilkan outbox event.
- Direct-dispatch bridge untuk transisi tanpa rewrite handler.

Exit criteria: satu scheduled job menghasilkan satu outbox event dan tidak ada
eksekusi handler langsung dari Scheduler Dispatcher.

### Phase 5: RabbitMQ consumer topology

Tujuan: menyediakan durable exchange, queue, retry queue, TTL, DLX, dan DLQ.

Pekerjaan:

- Tambahkan `github.com/rabbitmq/amqp091-go` ke `go.mod`.
- Tambahkan config RabbitMQ: DSN, exchange, DLQ exchange, queue role, prefetch, dan
  timeout.
- Tambahkan RabbitMQ durable service, volume, healthcheck, dan environment config ke
  `docker-compose.dev.yml` serta `.env.example`.
- Declare topic exchange `sipon.events` dan DLX `sipon.events.dlx`.
- Declare main queue per consumer role.
- Declare retry queue dengan TTL bertingkat dan dead-letter routing kembali ke main
  exchange.
- Declare DLQ per consumer role.
- Pastikan topology dibuat oleh adapter RabbitMQ, bukan oleh module handler.

Deliverables:

- Topology `sipon.worker.scheduler`, retry queue, dan DLQ.
- Management UI dapat melihat queue, consumer, retry, dan DLQ.

Exit criteria: RabbitMQ restart tidak menghilangkan durable topology dan test topology
berhasil dijalankan dari clean broker.

### Phase 6: Outbox Relay/Publisher

Tujuan: mengirim `event_outbox` yang sudah committed ke RabbitMQ secara reliable.

Pekerjaan:

- Implementasikan claim batch outbox `PENDING` atau retryable `FAILED`.
- Tambahkan status `PUBLISHING`, lease, attempt count, dan next attempt time.
- Implementasikan publisher confirm.
- Implementasikan connection/channel lifecycle dan reconnect.
- Tandai outbox `PUBLISHED` hanya setelah broker confirm.
- Tambahkan outbox retry dengan backoff; jangan menghapus event yang gagal.
- Serialisasi publish per channel atau gunakan channel pool yang aman.

Deliverables:

- `Outbox Relay/Publisher` dengan structured logging.
- Test broker down, reconnect, confirm failure, duplicate publish, dan lease recovery.

Exit criteria: event committed di PostgreSQL tetap dapat dipublish setelah RabbitMQ
sempat down.

### Phase 7: Message Consumer and durable inbox

Tujuan: memastikan setiap message memiliki record durable sebelum handler melakukan
side effect.

Pekerjaan:

- Consume message dengan manual acknowledgement.
- Parse envelope dan buat `message_jobs` dengan `ON CONFLICT DO NOTHING`.
- Commit inbox record berstatus `PENDING` sebelum handler dipanggil.
- Claim inbox row menjadi `RUNNING` dengan lease.
- Jika `SUCCEEDED`, ack tanpa memanggil handler lagi.
- Jika `FAILED`, ack dan menunggu replay operasional eksplisit.
- Dispatch ke shared registry dan module handler.
- Saat sukses, update inbox menjadi `SUCCEEDED` sebelum ack.
- Recover panic per message dan simpan error terakhir.

Deliverables:

- `Message Consumer` dengan alur `RabbitMQ -> message_jobs -> Module Handler`.
- Durable inbox adapter dan idempotency guard.

Exit criteria: duplicate delivery tidak menjalankan side effect kedua kali dan message
tidak hilang jika consumer crash sebelum acknowledgement.

### Phase 8: Retry and DLQ behavior

Tujuan: menghubungkan error classification handler dengan consumer topology.

Pekerjaan:

- `FatalError` mengubah inbox menjadi `FAILED` dan mengarahkan message ke DLQ.
- Error retryable dengan attempt tersedia mengirim message ke retry queue TTL.
- Setelah retry publish ter-confirm, update inbox menjadi `RETRY_WAIT` dan ack message
  asli.
- Setelah TTL habis, message kembali ke main queue dengan message ID yang sama.
- Retry berikutnya memakai row `message_jobs` yang sama, bukan membuat job ID baru.
- Setelah max attempts, state menjadi `FAILED` dan message berakhir di DLQ.
- Unknown routing key dan payload invalid diperlakukan sebagai fatal.

Deliverables:

- Retry policy per routing key.
- State transition lengkap `PENDING`, `RUNNING`, `RETRY_WAIT`, `SUCCEEDED`, `FAILED`.
- Test transient error, fatal error, retry exhaustion, TTL delay, dan DLQ.

Exit criteria: tidak ada `Nack(requeue=true)` tanpa delay dan tanpa batas attempt.

### Phase 9: Akademik end-to-end pilot

Tujuan: memvalidasi seluruh alur menggunakan handler akademik yang sudah ada.

Pekerjaan:

- Register route akademik ke `sipon.worker.scheduler`.
- Jalankan alur `OpenSession -> scheduled_jobs -> Scheduler Dispatcher -> event_outbox`.
- Jalankan `Outbox Relay/Publisher -> RabbitMQ -> message_jobs`.
- Jalankan `message_jobs -> akademik/interfaces/mq -> application usecase`.
- Pastikan auto-close mem-pause recurring fingerprint job lewat application layer.
- Pastikan message ID atau reference bisnis dipakai sebagai idempotency key.
- Jalankan pilot melalui direct-dispatch bridge, lalu RabbitMQ, dengan handler yang
  sama.

Test wajib:

- Sukses sync fingerprint.
- Sukses auto-close session.
- Payload invalid.
- Error transient dan retry.
- Error fatal dan DLQ.
- Duplicate delivery.
- Worker crash pada setiap boundary.
- RabbitMQ restart.

Exit criteria: akademik berpindah dari direct-dispatch ke RabbitMQ tanpa mengubah
handler atau routing contract.

### Phase 10: Module migration

Tujuan: memigrasikan handler module lain setelah pilot stabil.

Pekerjaan:

- Migrasikan module berdasarkan prioritas dan risiko, satu module per iterasi.
- Setiap module wajib memiliki `interfaces/mq` dengan struktur yang sama.
- `cmd/worker` hanya melakukan composition dan tidak menyimpan routing key module.
- Tambahkan payload contract test dan handler idempotency test per module.
- Hapus `RegisterSchedulerHandlers` setelah module selesai migrasi.

Exit criteria: semua asynchronous handler terdaftar melalui
`RegisterMessageHandlers`, tanpa registry scheduler lama.

### Phase 11: Observability, operations, and rollout

Tujuan: memastikan queue siap dijalankan dan dioperasikan di environment nyata.

Pekerjaan:

- Tambahkan log fields `message_id`, `routing_key`, `queue`, `correlation_id`,
  `attempt`, `duration`, `status`, dan `error_class`.
- Tambahkan metric outbox pending, publish failure, consumer throughput, handler
  duration, retry count, DLQ count, oldest pending age, dan stuck inbox count.
- Tambahkan health check PostgreSQL dan RabbitMQ.
- Buat command replay DLQ yang eksplisit dan memiliki audit log.
- Buat prosedur drain queue dan graceful shutdown dengan wait group.
- Tambahkan alert outbox tertua, lease expired, consumer stopped, dan DLQ bertambah.
- Aktifkan feature flag per routing key atau module.
- Bandingkan latency, success rate, retry rate, dan duplicate rate dengan baseline.
- Hapus direct-dispatch bridge dan stub `eventbus` setelah cutover selesai.
- Perbarui `docs/plan/worker-scheduler-architecture.md` agar menjelaskan arsitektur
  final.

Exit criteria: seluruh module berjalan melalui RabbitMQ, direct-dispatch bridge sudah
dihapus, dan operational checks lulus.

## File Area Yang Akan Terlibat

### Tambah

- `internal/shared/messaging/` untuk envelope, port, registry, error, dan policy.
- `internal/shared/messaging/infrastructure/rabbitmq/` untuk AMQP adapter.
- `internal/shared/messaging/infrastructure/persistence/` untuk outbox dan message job.
- `migrations/<timestamp>_create_event_outbox.up.sql` dan `.down.sql`.
- `migrations/<timestamp>_create_message_jobs.up.sql` dan `.down.sql`.
- `migrations/<timestamp>_add_scheduler_leases.up.sql` dan `.down.sql` bila lease
  belum ditambahkan ke migration scheduler.
- Test unit/integration untuk publisher, relay, lifecycle, retry, dan idempotency.

### Ubah

- `go.mod` dan `go.sum`.
- `internal/shared/config/config.go`.
- `cmd/worker/main.go` untuk wiring publisher, relay, consumer, dan shutdown.
- `cmd/api/main.go` hanya bila usecase membutuhkan outbox/transactor dependency.
- `internal/shared/scheduler/application/worker.go`.
- `internal/shared/scheduler/infrastructure/persistence/postgres_scheduled_job_repository.go`.
- `docker-compose.dev.yml` dan `.env.example`.
- `internal/modules/*/module.go` untuk facade `RegisterMessageHandlers`.
- `internal/modules/*/interfaces/mq/` untuk routing, payload, handler, dan registration.
- `internal/modules/akademik/application/command/job_types.go` untuk penghapusan
  constant routing lama dan injeksi route configuration.
- `docs/plan/worker-scheduler-architecture.md` setelah implementasi final.

## Reliability Rules

- At-least-once delivery adalah guarantee yang realistis; exactly-once tidak dijanjikan.
- Semua handler yang memiliki side effect wajib idempotent.
- Ack hanya dilakukan setelah status sukses atau dead-letter tercatat.
- Outbox tidak boleh dihapus ketika publish gagal.
- Message tidak boleh di-requeue tanpa delay dan tanpa batas attempt.
- Unknown routing key diperlakukan sebagai fatal dan masuk DLQ, bukan retry selamanya.
- Payload invalid diperlakukan sebagai fatal setelah envelope berhasil dicatat.
- Lease timeout harus lebih panjang dari timeout handler maksimum yang diizinkan.
- Shutdown menghentikan claim baru, memberi waktu handler aktif selesai, lalu menutup
  consumer dan publisher.

## Verification Matrix

### Build and static checks

- `go build ./...`
- `go test ./...`
- `go vet ./...`

### Database and outbox

- Business transaction rollback tidak meninggalkan outbox.
- Business commit selalu menghasilkan outbox.
- Dua relay worker tidak mengambil row outbox yang sama.
- Crash pada status `PUBLISHING` dapat dipulihkan setelah lease timeout.
- Publish confirm gagal membuat outbox retryable, bukan `PUBLISHED`.

### Scheduler

- One-off job menghasilkan tepat satu outbox event.
- Recurring job menghitung jadwal berikutnya tanpa melewatkan state.
- Job `PROCESSING` yang lease-nya expired dapat diambil worker lain.
- Error relay tidak memanggil handler secara langsung.

### Consumer and RabbitMQ

- Message sukses di-ack dan menjadi `SUCCEEDED`.
- Duplicate message ID tidak mengeksekusi side effect kedua kali.
- Error transient masuk retry queue dengan delay yang bertambah.
- Error fatal langsung masuk DLQ dan menjadi `FAILED`.
- Retry yang melewati batas tidak berputar di queue utama.
- RabbitMQ restart menyebabkan reconnect dan message durable tetap tersedia.
- SIGTERM menghentikan consumer tanpa kehilangan message yang belum di-ack.

### Operational checks

- Queue, retry queue, dan DLQ terlihat dari RabbitMQ management UI.
- Log dapat menghubungkan HTTP request, outbox row, message, dan handler lewat
  `correlation_id`.
- Ada cara replay satu message DLQ tanpa mengubah payload secara manual di database.
- Ada alert atau query untuk outbox tertua, message stuck, dan DLQ bertambah.

## Acceptance Criteria

Implementasi dianggap selesai bila:

1. `scheduled_jobs` tidak lagi menjalankan handler bisnis secara langsung.
2. Perubahan bisnis dan event outbox tersimpan dalam satu transaksi.
3. Worker dapat berjalan lebih dari satu instance tanpa double claim.
4. RabbitMQ publish memakai confirm dan outbox tidak kehilangan event saat broker
   gagal.
5. Consumer memiliki idempotency, bounded retry, delay, dan DLQ.
6. Worker dapat graceful shutdown dan reconnect ke RabbitMQ.
7. Semua module mendaftarkan handler melalui contract `RegisterMessageHandlers` yang
   seragam; tidak ada `RegisterSchedulerHandlers` tersisa.
8. Akademik dapat dipindahkan dari direct-dispatch bridge ke RabbitMQ tanpa mengubah
   handler atau routing contract.
9. Pilot module lulus verification matrix dan seluruh package lulus build/test.
10. Dokumentasi arsitektur, konfigurasi, routing key, operasi retry, dan DLQ tersedia
   di `docs`.

## Risiko dan Mitigasi

| Risiko | Mitigasi |
| --- | --- |
| Duplicate delivery setelah crash | `message_jobs` + idempotent handler + unique business constraint |
| Outbox membesar | Index claim, batch limit, retention job, dan alert oldest age |
| Retry hot-loop | Retry queues TTL, backoff, jitter, dan max attempts |
| RabbitMQ connection putus | Connection manager dengan reconnect dan publish confirm |
| Handler terlalu lama | Context timeout, lease yang sesuai, dan queue role terpisah |
| Schema envelope berubah | `version`, contract test, dan routing key baru untuk breaking change |
| DLQ diabaikan | Metric, alert, dashboard, dan prosedur replay yang terdokumentasi |
| Migrasi terlalu besar | Pilot satu routing key sebelum migrasi seluruh module |
