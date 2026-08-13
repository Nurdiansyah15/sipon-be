# Arsitektur Worker/Scheduler (General-Purpose)

## Ringkasan

Sistem background job berbasis **database-persisted scheduled jobs** yang dijalankan oleh
**worker process terpisah**. Usecase mendefinisikan job schedule → persist ke DB → worker
tick secara recurring, mengklaim dan mengeksekusi job yang jatuh tempo.

Terinspirasi dari pola yang diterapkan di `k-forum-api` (`app/service/scheduledjob`),
disesuaikan tanpa RabbitMQ — worker langsung mengeksekusi handler via registry dispatch.

## Arsitektur

```
┌──────────────────────┐       ┌─────────────────────┐
│   cmd/api (HTTP)     │       │   cmd/worker        │
│                      │       │                     │
│  Usecase menulis:    │       │  Worker tick loop:  │
│  - scheduled_jobs    │       │  1. FindDueAndClaim │
│                      │       │     (10s interval)  │
│  Tidak ada ticker/   │       │  2. Dispatch via    │
│  polling di proses   │       │     Registry →      │
│  API.                │       │     HandlerFunc     │
└──────────┬───────────┘       └──────────┬──────────┘
           │                              │
           ▼                              ▼
    ┌──────────────────────────────────────────┐
    │           PostgreSQL                      │
    │  scheduled_jobs table                    │
    │  (FOR UPDATE SKIP LOCKED)                │
    └──────────────────────────────────────────┘
```

## Alur Generic

1. **Usecase** (di `cmd/api`) membuat `ScheduledJob` (one-off atau recurring) dan persist
   ke tabel `scheduled_jobs` via `scheduler.Repository.Save()`.
2. **Worker** (proses terpisah, `cmd/worker`) menjalankan ticker loop setiap N detik.
3. Setiap tick, worker mengklaim job yang `status = 'ACTIVE' AND next_run_at <= now`
   menggunakan `SELECT ... FOR UPDATE SKIP LOCKED` — aman multi-instance.
4. Job yang diklaim di-set `status = 'PROCESSING'`, lalu handler dipanggil via
   `Registry.Dispatch(ctx, jobType, payload)`.
5. Setelah eksekusi:
   - **Recurring**: `MarkFired()` → hitung `next_run_at` berikutnya dari cron expr → `status = 'ACTIVE'`
   - **One-off**: `MarkCompleted()` → `status = 'COMPLETED'`
   - **Error**: `MarkFailed()` → retry jika `retry_count < max_retry`,否则 `status = 'FAILED'`

## Struktur File

```
internal/shared/scheduler/
├── domain/
│   └── scheduled_job/
│       ├── constant/
│       │   └── constant.go                          # ScheduleType, Status enums
│       ├── entity/
│       │   └── scheduled_job.go                     # Aggregate root + domain methods
│       └── repository/
│           └── interfaces.go                        # Repository port
├── application/
│   ├── errors.go                                    # ErrHandlerNotFound
│   ├── registry.go                                  # Registry, HandlerFunc, RetryPolicy, error types
│   └── worker.go                                    # Worker (ticker loop, processBatch, executeJob)
└── infrastructure/
    └── persistence/
        └── postgres_scheduled_job_repository.go      # Postgres adapter

cmd/worker/
└── main.go                                          # Entrypoint worker process

migrations/
└── 20260813000000_create_scheduled_jobs.up.sql
```

## Database Schema

```sql
CREATE TABLE scheduled_jobs (
    id            UUID         PRIMARY KEY,
    type          VARCHAR(100) NOT NULL,        -- routing key (mis. "fingerprint.sync_open_sessions")
    payload       JSONB        NOT NULL DEFAULT '{}',
    schedule_type VARCHAR(20)  NOT NULL CHECK (schedule_type IN ('ONE_OFF', 'RECURRING')),
    cron_expr     VARCHAR(100),                 -- NULL untuk ONE_OFF
    run_at        TIMESTAMPTZ,                  -- NULL untuk RECURRING
    next_run_at   TIMESTAMPTZ  NOT NULL,
    last_run_at   TIMESTAMPTZ,
    status        VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE'
                  CHECK (status IN ('ACTIVE','PROCESSING','PAUSED','COMPLETED','FAILED')),
    retry_count   INT          NOT NULL DEFAULT 0,
    max_retry     INT          NOT NULL DEFAULT 3,
    last_error    TEXT,
    reference_id  VARCHAR(255),                 -- opaque FK ke entitas bisnis
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scheduled_jobs_due      ON scheduled_jobs (next_run_at, status) WHERE status = 'ACTIVE';
CREATE INDEX idx_scheduled_jobs_type_ref ON scheduled_jobs (type, reference_id)  WHERE reference_id IS NOT NULL;
```

## Domain Model — `domain/scheduled_job/entity/scheduled_job.go`

```go
type ScheduledJob struct {
    ID           uuid.UUID
    Type         string
    Payload      json.RawMessage
    ScheduleType constant.ScheduleType    // ONE_OFF | RECURRING
    CronExpr     *string                  // nil untuk ONE_OFF
    RunAt        *time.Time               // nil untuk RECURRING
    NextRunAt    time.Time
    LastRunAt    *time.Time
    Status       constant.Status          // ACTIVE | PROCESSING | PAUSED | COMPLETED | FAILED
    RetryCount   int
    MaxRetry     int
    LastError    *string
    ReferenceID  *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

func NewOneOffJob(jobType string, payload json.RawMessage, runAt time.Time) *ScheduledJob
func NewRecurringJob(jobType string, payload json.RawMessage, cronExpr string,
                     parser cron.Parser, loc *time.Location) (*ScheduledJob, error)

func (j *ScheduledJob) MarkFired(nextRunAt time.Time)    // recurring: advance NextRunAt
func (j *ScheduledJob) MarkCompleted()                    // one-off: terminal
func (j *ScheduledJob) MarkFailed(errMsg string)          // increment retry, set FAILED jika max tercapai
func (j *ScheduledJob) Pause()
func (j *ScheduledJob) Resume()
func (j *ScheduledJob) UpdateSchedule(cronExpr string, parser cron.Parser, loc *time.Location) error
```

## Registry & Handler — `application/registry.go`

```go
type HandlerFunc func(ctx context.Context, payload json.RawMessage) error

type Registry struct { ... }
func NewRegistry() *Registry
func (r *Registry) Register(jobType string, handler HandlerFunc)
func (r *Registry) Dispatch(ctx context.Context, jobType string, payload json.RawMessage) error

// Error types untuk signaling retry behavior
type RetryableError struct{ Err error }  // akan di-retry
type FatalError     struct{ Err error }  // langsung FAILED, no retry
func IsFatal(err error) bool

// RetryPolicy: configurable max retry per job type
type RetryPolicy struct { ... }
func NewRetryPolicy(defaultMax int) *RetryPolicy
func (p *RetryPolicy) Register(jobType string, maxRetry int)
func (p *RetryPolicy) MaxRetryFor(jobType string) int
```

## Worker — `application/worker.go`

```go
type Worker struct { ... }
func NewWorker(repo repository.Repository, registry *Registry, tick time.Duration, logger *slog.Logger) *Worker
func (w *Worker) Run(ctx context.Context)
```

`Run()` menjalankan ticker loop:
- Setiap tick: `FindDueAndClaim()` → klaim batch (max 50) via `FOR UPDATE SKIP LOCKED`
- Tiap job: `Registry.Dispatch()` → handler execution
- Error handling: `recover()` per-job, error di-log, job state di-update
- Context cancellation → graceful stop

## Repository — `domain/scheduled_job/repository/interfaces.go`

```go
type Repository interface {
    Save(ctx context.Context, job *entity.ScheduledJob) error
    FindDueAndClaim(ctx context.Context, now time.Time, limit int) ([]*entity.ScheduledJob, error)
    Update(ctx context.Context, job *entity.ScheduledJob) error
    FindByTypeAndReferenceID(ctx context.Context, jobType string, referenceID string) (*entity.ScheduledJob, error)
}
```

`FindDueAndClaim` — dua fase dalam satu transaksi:
1. `SELECT ... WHERE status='ACTIVE' AND next_run_at <= $1 FOR UPDATE SKIP LOCKED LIMIT $2`
2. `UPDATE SET status='PROCESSING' WHERE id = ANY(claimed_ids)`
3. COMMIT → row locks released, jobs dieksekusi di luar transaksi

Implementasi: `infrastructure/persistence/postgres_scheduled_job_repository.go`

## Module Integration

Setiap module expose method `RegisterSchedulerHandlers(registry *application.Registry)` di
`module.go`. Method ini disebut oleh `cmd/worker/main.go` saat wiring. Module mendaftarkan
handler untuk job type yang menjadi tanggung jawabnya.

```go
func (m *Module) RegisterSchedulerHandlers(registry *application.Registry) {
    registry.Register("fingerprint.sync_open_sessions", func(ctx context.Context, payload json.RawMessage) error {
        // parse payload, panggil use case
        return m.syncOpenSessionsUC.Execute(ctx, ...)
    })
}
```

## Concurrency Safety

- **`FOR UPDATE SKIP LOCKED`**: Multi-worker instance dapat berjalan bersamaan tanpa
  double-fire. Worker yang terlambat akan skip row yang sudah dikunci.
- **Status machine**: `ACTIVE → PROCESSING → ACTIVE/COMPLETED/FAILED` — row lock
  mencegah race condition saat klaim.
- **Crash recovery**: Job yang stuck di `PROCESSING` (worker crash saat eksekusi)
  dapat di-recover oleh monitoring/manual intervention. Future: tambahkan
  `PROCESSING` timeout detection.

## Config — `internal/shared/config/config.go`

```go
type WorkerConfig struct {
    Enabled     bool
    TickSeconds int
}
// di Config: Worker WorkerConfig
// di Load(): Worker: WorkerConfig{
//     Enabled:     getEnv("WORKER_ENABLED", "true") != "false",
//     TickSeconds: parseInt("WORKER_TICK_SECONDS", 10),
// }
```

`WORKER_ENABLED=false` untuk development/debugging tanpa background job.
`WORKER_TICK_SECONDS` mengatur interval polling (default 10 detik).

## Docker Compose — Worker Service

```yaml
worker:
  image: golang:1.25-alpine
  container_name: sipon-be_worker
  restart: unless-stopped
  env_file: .env
  working_dir: /workspace
  command: ["go", "run", "./cmd/worker"]
  volumes:
    - .:/workspace
    - go_mod_cache:/go/pkg/mod
    - go_build_cache:/root/.cache/go-build
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
  networks:
    - sipon-net
```

`cmd/api` dan `cmd/worker` adalah dua proses terpisah dengan entry point berbeda,
menggunakan database yang sama. API tidak menjalankan ticker/scheduler apapun.

## cmd/worker/main.go

Worker entry point:
1. Load config, init timezone, logger
2. Connect PostgreSQL & Redis
3. Construct semua module (sama seperti `cmd/api`)
4. Buat `application.Registry`, panggil `RegisterSchedulerHandlers()` di setiap module
5. Buat `application.NewWorker()`, panggil `worker.Run(ctx)`
6. Graceful shutdown: listen SIGINT/SIGTERM → cancel context → wait

## Perbedaan dengan k-forum-api

| Aspek | k-forum-api | sipon-be |
|-------|-------------|----------|
| Message queue | RabbitMQ (outbox → MQ → consumer) | **Tanpa MQ** — worker langsung execute |
| Delivery | Two-phase: outbox relay + scheduler relay | **Single-phase**: worker langsung dispatch |
| Error handling | MQ nack/requeue + DLQ | Retry count + status machine di DB |
| Concurrency | FOR UPDATE SKIP LOCKED + temp hold | FOR UPDATE SKIP LOCKED + status transition |
| Cron parsing | robfig/cron/v3 | robfig/cron/v3 (sama) |
| Job model | scheduled_jobs + jobs + event_outbox | **scheduled_jobs saja** |

## Perbedaan dengan Plan Sebelumnya

| Aspek | Plan Lama (in-process) | Plan Baru (separate worker) |
|-------|----------------------|---------------------------|
| Lokasi eksekusi | Di proses API (`cmd/api`) | Proses terpisah (`cmd/worker`) |
| Job definition | Hardcoded di `main.go` via `Register()` | **Persisted di DB** oleh usecase |
| Storage | In-memory entries | **PostgreSQL** `scheduled_jobs` table |
| Concurrency | Redis distributed lock (optional) | **FOR UPDATE SKIP LOCKED** (built-in) |
| Scalability | Single instance | **Multi-worker** safe |
| Job lifecycle | Ticker per goroutine | **Centralized** ticker loop |
| Retry | Tidak ada | **Retry count** + max retry per type |

## Cara Menambahkan Job Baru

1. **Definisikan routing key** di module (misal `fingerprint/job_types.go`):
   ```go
   const RoutingSyncOpenSessions = "fingerprint.sync_open_sessions"
   ```

2. **Daftarkan handler** di `module.go`:
   ```go
   func (m *Module) RegisterSchedulerHandlers(registry *scheduler.Registry) {
       registry.Register(RoutingSyncOpenSessions, m.handleSyncOpenSessions)
   }
   func (m *Module) handleSyncOpenSessions(ctx context.Context, payload json.RawMessage) error {
       var p SyncPayload
       json.Unmarshal(payload, &p)
       return m.syncUC.Execute(ctx, p.SessionID)
   }
   ```

3. **Persist job** dari usecase:
   ```go
   // One-off: jalankan 5 menit dari sekarang
   job := scheduler.NewOneOffJob(RoutingSyncOpenSessions, payload, time.Now().Add(5*time.Minute))
   scheduledJobRepo.Save(ctx, job)

   // Recurring: setiap 2 menit
   job, _ := scheduler.NewRecurringJob(RoutingSyncOpenSessions, payload, "*/2 * * * *",
       cron.NewParser(cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow),
       timeutil.Loc())
   scheduledJobRepo.Save(ctx, job)
   ```

## Verifikasi

1. `go build ./...` — semua package termasuk `cmd/worker` compile
2. `go run ./cmd/worker` — worker start, log "scheduler worker started" muncul
3. Insert manual ke `scheduled_jobs` dengan `next_run_at` di masa lalu:
   ```sql
   INSERT INTO scheduled_jobs (id, type, payload, schedule_type, next_run_at, status)
   VALUES (gen_random_uuid(), 'test.echo', '{}', 'ONE_OFF', now() - interval '1 minute', 'ACTIVE');
   ```
   Pastikan worker log eksekusi job tersebut
4. Kirim SIGTERM ke worker — pastikan graceful stop tanpa panic
5. `WORKER_ENABLED=false` di `.env` — worker tetap start tapi tidak ada job terdaftar
6. Multi-instance: jalankan 2 worker sekaligus, pastikan job tidak dieksekusi dobel

## Dependencies

- `github.com/robfig/cron/v3` — cron expression parsing (sudah ditambahkan ke go.mod)
- PostgreSQL — job persistence + row-level locking
- Existing `internal/shared/` packages (config, database, logger, timeutil)

## Catatan Implementasi

- `RegisterSchedulerHandlers` saat ini **no-op** di semua module — siap diisi handler
  sesuai kebutuhan (lihat `docs/plan/fingerprint-attendance-integration.md` untuk
  consumer pertama yang direncanakan).
- `timeutil.Loc()` dipakai sebagai timezone untuk evaluasi cron expression, konsisten
  dengan platform timezone yang sudah ada.
- `reference_id` adalah opaque string — bisa dipakai untuk dedup/cancel job
  (misal: "the scrape job for source X") via `FindByTypeAndReferenceID`.
- Struktur folder mengikuti konvensi DDD yang sama dengan module:
  `domain/` (entity, constant, repository port), `application/` (worker, registry, errors),
  `infrastructure/persistence/` (postgres adapter).
