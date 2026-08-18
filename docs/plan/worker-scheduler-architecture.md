# Arsitektur Scheduler & Messaging Queue

Dokumen ini menjelaskan arsitektur **final** background job dan event delivery
setelah penyempurnaan messaging queue. Rincian fase dan keputusan ada di
`docs/plan/messaging-queue-infrastructure.md`.

## Ringkasan

Sistem memisahkan tiga tanggung jawab:

1. **Scheduling** — kapan pekerjaan harus dipicu (`scheduled_jobs`).
2. **Reliable publish** — perubahan bisnis + `event_outbox` dalam satu transaksi,
   lalu dibawa ke RabbitMQ.
3. **Eksekusi handler** — consume dari RabbitMQ lewat durable inbox
   (`message_jobs`), dispatch ke handler module.

## Alur

```
HTTP API / Module Usecase
        │  DB Transaction: business write + event_outbox INSERT
        ▼
   PostgreSQL
        │  Outbox Relay (claim PENDING, publisher confirm)
        ▼
RabbitMQ Topic Exchange  (sipon.events)
        │
        ▼  queue per consumer role (sipon.worker.scheduler, ...)
        +-- Retry Queue (TTL, x-dead-letter-routing-key = routing key asli)
        +-- DLQ
        │
        ▼
Message Consumer: RabbitMQ -> message_jobs (durable inbox) -> Module Handler
```

Worker process menjalankan:
- **Scheduler Dispatcher** — claim `scheduled_jobs` jatuh tempo (lease recovery
  untuk `PROCESSING` yang stuck), lalu tulis `event_outbox` + update state job
  dalam satu transaksi. Tidak pernah memanggil handler secara langsung.
- **Outbox Relay** — claim `event_outbox` `PENDING`/retryable `FAILED`, publish ke
  RabbitMQ dengan publisher confirm; `PUBLISHED` hanya setelah broker confirm.
- **Message Consumer** — alur `RabbitMQ -> message_jobs -> handler`, dengan
  idempotency, retry TTL, dan DLQ.

## Komponen Utama

### Shared messaging (`internal/shared/messaging`)

- `Message` / `Binding` / `Registry` / `RetryPolicy` — kontrak transport.
- `domain/event_outbox` — entity + repository port.
- `domain/message_job` — durable inbox entity + repository port.
- `application/` — `OutboxRelay`, `MessageConsumer`, `Metrics`.
- `infrastructure/rabbitmq/` — `Topology`, `RabbitMQPublisher`, `RabbitMQConsumer`.
- `infrastructure/persistence/` — adapter Postgres outbox + message_jobs.

### Scheduler (`internal/shared/scheduler`)

- `application/dispatcher.go` — `SchedulerDispatcher` (mode direct/outbox).
- `application/errors.go` — `FatalError` / `RetryableError` / `IsFatal`.
- `domain/scheduled_job` + `infrastructure/persistence` — lease-based claim.

### Module

- Setiap module meng-expose `RegisterMessageHandlers(registry) ([]Binding, error)`.
- Handler, payload, routing key, dan binding berada di
  `internal/modules/<module>/interfaces/mq/`.

## State Machine

`message_jobs`: `PENDING -> RUNNING -> SUCCEEDED`, `RUNNING -> RETRY_WAIT ->
RUNNING`, `RUNNING -> FAILED`.

`event_outbox`: `PENDING -> PUBLISHING -> PUBLISHED`, gagal -> `FAILED` (retry
dengan backoff).

`scheduled_jobs`: `ACTIVE -> PROCESSING -> ACTIVE/COMPLETED/FAILED` + lease.

## Retry & DLQ

- Retryable error, attempt tersedia: publish ke retry queue TTL, tandai
  `RETRY_WAIT`, ack pesan asli. Setelah TTL, message kembali ke main queue dengan
  message ID dan routing key yang sama.
- Fatal atau retry habis: `FAILED` dan nack tanpa requeue -> DLQ.
- Tidak ada `Nack(requeue=true)` tanpa delay dan tanpa batas attempt.

## Observability & Ops

- Structured log: `id`, `routing_key`, `correlation_id`, `attempt`, `duration`,
  `status`, `error_class`.
- Health server worker: `/healthz` (PostgreSQL + RabbitMQ) dan `/metrics`
  (counter outbox/consumer) pada `WORKER_HEALTH_PORT`.
- Graceful shutdown via WaitGroup (bukan `time.Sleep`).
- Lease recovery menangani worker crash; topology durable terhadap restart broker.

## Konfigurasi

- `WORKER_MODE` = `direct` (compatibility) atau `outbox` (pipeline penuh).
- `WORKER_TICK_SECONDS`, `WORKER_LEASE_SECONDS`, `WORKER_HEALTH_PORT`.
- `RABBITMQ_*` (DSN, exchange, DLX, prefetch, timeout, retry delays, enabled).
- Aktifkan pipeline penuh: `RABBITMQ_ENABLED=true` + `WORKER_MODE=outbox`.

## Cara Menambahkan Job Baru

1. Definisikan routing key dan binding di `interfaces/mq/routing.go`.
2. Definisikan payload DTO di `interfaces/mq/payload.go`.
3. Tulis handler adapter di `interfaces/mq/handler.go` (decode -> call usecase ->
   classify error).
4. Daftarkan di `interfaces/mq/register.go` dan expose lewat
   `Module.RegisterMessageHandlers`.
5. `cmd/worker` otomatis memakai binding yang dikembalikan.
