# Plan: Arsitektur Worker/Scheduler (General-Purpose)

## Context

Codebase ini saat ini **tidak punya** infrastruktur background job / scheduler sama sekali.
Satu-satunya goroutine yang ada: `go func() { srv.ListenAndServe() }()` di `cmd/api/main.go:165-172`
(menjalankan HTTP server) dan satu goroutine fire-and-forget di
`internal/modules/feedback/application/command/delete_feedback.go:69` (cleanup sekali jalan,
bukan recurring). Tidak ada `cron`/`ticker`/worker package di manapun.

Kebutuhan untuk pekerjaan berkala (recurring) sudah mulai muncul dan akan makin sering, contoh
konkret yang sudah teridentifikasi:
- Sync data dari mesin fingerprint untuk sesi yang sedang `open` (lihat
  `docs/plan/fingerprint-attendance-integration.md` — di plan itu sengaja **manual trigger**
  dulu karena infra ini belum ada; begitu ada, auto-poll bisa jadi consumer pertama).
- Auto-close sesi/periode yang lewat waktu, reminder herregistrasi mendekati deadline, cleanup
  data kadaluarsa, dll — semua pola "jalankan fungsi X setiap interval Y" yang berulang di
  banyak module.

Daripada setiap module bikin ticker-nya sendiri-sendiri (duplikasi pola, tidak ada shutdown
yang konsisten, tidak ada observability yang seragam), plan ini membuat **satu package
shared** yang generik dan dipakai module manapun, mengikuti konvensi yang sudah ada di
`internal/shared/` (`database`, `middleware`, `logger`, dst: struct + constructor sederhana,
**tanpa** DI container/registry — lihat `docs/architecture/module-boundaries.md` baris
173-175 yang secara eksplisit melarang itu).

## Keputusan Desain

| No | Pertanyaan | Keputusan | Alasan |
|----|-----------|-----------|--------|
| 1 | Cron-expression (`"0 6 * * *"`) atau interval sederhana (`time.Duration`)? | **Interval sederhana** (`5*time.Minute`, dst), tiap job juga bisa `RunOnStart bool`. | YAGNI — belum ada kebutuhan nyata untuk cron-expression; kalau nanti perlu, tinggal tambah field `Schedule` opsional tanpa mengubah interface `Job`. Menambah dependency (`robfig/cron`) sekarang cuma buat kebutuhan hipotetis melanggar prinsip yang sudah dipegang repo ini. |
| 2 | Perlu distributed lock (biar job tidak dobel-run kalau ada >1 instance)? | **Ya, tapi opsional/nil-safe.** `Locker` adalah interface; scheduler jalan normal tanpa lock kalau `nil` (cocok untuk dev/single-instance). Implementasi default: `RedisLocker` (SET NX + TTL) karena Redis sudah ada & wired (`cmd/api/main.go:53-58`). | Topologi produksi saat ini tidak terdeklarasi di repo (`docker-compose.dev.yml` cuma 1 service app, tidak ada k8s manifest) — desain harus aman untuk kedua skenario tanpa memaksa Redis jadi hard dependency scheduler. |
| 3 | Bagaimana module "mendaftarkan" job tanpa registry generik? | **`main.go` yang merakit**, sama seperti semua wiring lain di repo ini. Module expose method public biasa (mis. `func (m *Module) SyncOpenSessionsFingerprint(ctx) error`) — bukan lewat `Contract` (itu untuk panggilan *antar-module*, bukan main.go memanggil job miliknya sendiri) — lalu `main.go` bungkus jadi `worker.JobFunc{...}` dan `scheduler.Register(...)`. | Konsisten dengan larangan DI container/registry di `module-boundaries.md`; `main.go` sudah jadi satu-satunya tempat yang "tahu semua module secara konkret", jadi wajar dia juga jadi tempat merakit job. |
| 4 | Error di satu job bikin proses lain ikut down? | **Tidak.** Tiap job jalan di goroutine sendiri, `recover()` per-eksekusi, error di-log via `slog` tapi scheduler & job lain tetap jalan. | Satu job yang salah (mis. koneksi eksternal fingerprint device timeout) tidak boleh mematikan server API. |
| 5 | Bagaimana graceful shutdown? | Scheduler menerima `context.Context` yang di-cancel di titik yang sama dengan shutdown server (`cmd/api/main.go` baris `<-quit` s.d. `srv.Shutdown`), lalu `scheduler.Stop(timeout)` menunggu job yang sedang berjalan selesai (dibatasi timeout, sama gaya `srv.Shutdown(ctx)` yang sudah pakai `10*time.Second`). | Meniru pola shutdown yang sudah ada, bukan pola baru. |

## Komponen

### `internal/shared/worker/job.go`

```go
package worker

import "context"

type Job interface {
    Name() string
    Run(ctx context.Context) error
}

// JobFunc membungkus fungsi biasa jadi Job, supaya module tidak perlu bikin
// struct baru cuma untuk register satu job (mirip http.HandlerFunc).
type JobFunc struct {
    JobName string
    Fn      func(ctx context.Context) error
}

func (f JobFunc) Name() string                    { return f.JobName }
func (f JobFunc) Run(ctx context.Context) error   { return f.Fn(ctx) }
```

### `internal/shared/worker/locker.go`

```go
package worker

import "context"
import "time"

// Locker mencegah job yang sama jalan bersamaan di >1 instance. Opsional —
// Scheduler jalan tanpa lock kalau Locker nil (single-instance/dev).
type Locker interface {
    // TryAcquire mencoba ambil lock untuk `key`. Kalau berhasil, `ok=true`
    // dan `release` harus dipanggil setelah job selesai. TTL adalah jaring
    // pengaman kalau proses crash sebelum release dipanggil.
    TryAcquire(ctx context.Context, key string, ttl time.Duration) (release func(), ok bool, err error)
}
```

### `internal/shared/worker/redis_locker.go`

```go
package worker

type RedisLocker struct {
    client *redis.Client
}

func NewRedisLocker(client *redis.Client) *RedisLocker { return &RedisLocker{client: client} }

func (l *RedisLocker) TryAcquire(ctx context.Context, key string, ttl time.Duration) (func(), bool, error) {
    ok, err := l.client.SetNX(ctx, "worker:lock:"+key, "1", ttl).Result()
    if err != nil || !ok {
        return nil, ok, err
    }
    release := func() { l.client.Del(context.Background(), "worker:lock:"+key) }
    return release, true, nil
}
```
Sama gaya dengan `internal/modules/identity/infrastructure/cache/redis_rate_limiter.go` (struct
tipis membungkus `*redis.Client`, method langsung pakai perintah Redis). **Catatan batasan**:
`Del` di sini tidak compare-and-delete (tidak cek token pemilik) — cukup untuk kasus TTL yang
jauh lebih besar dari durasi job normal; kalau nanti butuh jaminan lebih ketat, upgrade ke Lua
script CAS tanpa mengubah interface `Locker`.

### `internal/shared/worker/scheduler.go`

```go
type entry struct {
    job        Job
    interval   time.Duration
    runOnStart bool
}

type Scheduler struct {
    logger *slog.Logger
    locker Locker // boleh nil
    lockTTL time.Duration
    entries []entry
    wg      sync.WaitGroup
}

func NewScheduler(logger *slog.Logger, locker Locker) *Scheduler

// Register menambahkan job. Harus dipanggil sebelum Start.
func (s *Scheduler) Register(job Job, interval time.Duration, runOnStart bool)

// Start menjalankan semua job terdaftar, masing-masing di goroutine sendiri.
// Berhenti saat ctx di-cancel.
func (s *Scheduler) Start(ctx context.Context)

// Stop menunggu semua job yang sedang berjalan selesai, dibatasi timeout.
func (s *Scheduler) Stop(timeout time.Duration)
```

Isi `Start`: satu goroutine per entry, `select { case <-ctx.Done(): return; case <-ticker.C: s.execute(...) }`,
plus jalankan langsung sekali di awal kalau `runOnStart`. `execute` membungkus `job.Run` dengan
`recover()`, akuisisi `Locker` (kalau ada) sebelum run dan `release()` di akhir, dan log
durasi + hasil (`slog.Info`/`slog.Error`) — level `Info` untuk sukses, `Error` untuk gagal/panic,
konsisten dengan gaya logging di `cmd/api/main.go` (`lg.Info(...)`, `lg.Error(...)`).

### Config — `internal/shared/config/config.go`

```go
type WorkerConfig struct {
    Enabled bool
}
// di Config: Worker WorkerConfig
// di Load(): Worker: WorkerConfig{Enabled: getEnv("WORKER_ENABLED", "true") != "false"}
```
Satu flag global cukup untuk v1 (matikan semua job sekaligus, misal untuk debugging lokal).
Interval per-job **tidak** dikonfigurasi lewat env generik — itu keputusan tiap module saat
`Register` (hardcode default yang masuk akal, atau baca env spesifik module itu sendiri kalau
memang perlu dikonfigurasi, sama seperti `RateLimitConfig` per-domain yang sudah ada).

### Wiring — `cmd/api/main.go`

```go
scheduler := worker.NewScheduler(lg, worker.NewRedisLocker(redisClient))

if cfg.Worker.Enabled {
    // contoh consumer: auto-sync fingerprint untuk sesi yang open (opsional,
    // menyusul docs/plan/fingerprint-attendance-integration.md kalau mau
    // upgrade dari manual-trigger ke auto-poll)
    scheduler.Register(
        worker.JobFunc{JobName: "fingerprint-sync-open-sessions", Fn: akademik.SyncOpenSessionsFingerprint},
        2*time.Minute, false,
    )
}

workerCtx, cancelWorker := context.WithCancel(context.Background())
scheduler.Start(workerCtx)

// ... existing engine setup & srv.ListenAndServe() goroutine tidak berubah ...

<-quit
lg.Info("shutdown signal received")
cancelWorker()
scheduler.Stop(10 * time.Second)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

`akademik.SyncOpenSessionsFingerprint` (kalau dipakai) adalah method public baru di
`*akademikModule.Module` — bukan bagian dari `Contract` (itu untuk dipanggil module lain, bukan
`main.go` memanggil job miliknya sendiri) — isinya query semua sesi `status = open`
(`sessionRepo.List(ctx, ...)` yang sudah ada) lalu panggil
`SyncAttendanceFromFingerprintUseCase.Execute` per sesi.

## Pola Pakai untuk Module Lain

1. Module tidak perlu import `internal/shared/worker` di domain/application layer-nya — cukup
   punya satu method public di `*Module` yang isinya "jalankan use case X sekali", sama seperti
   method `RegisterRoutes`/`EnsurePendingUploadLifecycle` yang sudah ada.
2. `main.go` yang membungkusnya jadi `worker.JobFunc` dan `scheduler.Register(...)` dengan
   interval yang sesuai kebutuhan job itu.
3. Kalau job butuh lock antar-instance (mis. karena punya efek samping yang tidak idempoten),
   pastikan `key` yang dipakai di `Locker` unik per job (default: `job.Name()`).

## Fase Pengerjaan

1. **Package inti**: `internal/shared/worker/job.go` (`Job`, `JobFunc`), `locker.go` (interface),
   `redis_locker.go` (`RedisLocker`), `scheduler.go` (`Scheduler`, `Register`, `Start`, `Stop`).
   Unit test: job sukses ter-log, job error tidak menghentikan job lain, `runOnStart` jalan
   sekali di awal, `ctx` cancel menghentikan semua goroutine (pakai `interval` pendek + `context.WithTimeout` di test).
   `go build ./... && go test ./...`.
2. **Locker test**: unit test `RedisLocker` pakai `redis` test instance (miniredis atau redis
   container yang sudah dipakai test lain kalau ada) — dua `TryAcquire` dengan key sama, yang
   kedua harus `ok=false` sebelum TTL habis.
3. **Config + wiring dasar**: tambah `WorkerConfig` di `config.go`, buat `scheduler` di
   `main.go`, sambungkan `cancelWorker()` + `scheduler.Stop(...)` ke urutan shutdown yang sudah
   ada. **Belum ada job terdaftar** di fase ini — cuma infra kosong yang jalan & shutdown bersih.
   `go build ./...`, jalankan `go run ./cmd/api`, kirim `SIGTERM`, pastikan log shutdown rapi
   tanpa goroutine leak/panic.
4. **Consumer pertama (opsional, kalau mau langsung dipakai)**: tambah method
   `SyncOpenSessionsFingerprint` di akademik module (setelah plan fingerprint selesai
   diimplementasikan) dan register sebagai job — ini sekaligus jadi bukti integrasi end-to-end
   package baru ini.

## Verifikasi

1. `go build ./...` dan `go test ./...` lolos di tiap fase.
2. Jalankan app lokal dengan `WORKER_ENABLED=true` dan minimal satu job dummy (mis. job yang
   cuma `slog.Info("tick")`) dengan interval pendek (5s) — pastikan log muncul berkala.
3. Set `WORKER_ENABLED=false` — pastikan tidak ada job jalan (log tick tidak muncul), tapi
   server API tetap jalan normal.
4. Matikan Redis (stop container) sambil scheduler pakai `RedisLocker` — pastikan job tetap
   jalan (locker gagal acquire → `execute` skip run dengan log warning, bukan crash) **atau**
   sesuai desain, tentukan perilaku yang diinginkan (fail-open vs fail-closed) dan pastikan itu
   yang terjadi secara konsisten — dokumentasikan pilihannya di kode.
5. Kirim `SIGTERM` saat job sedang berjalan (job dummy dengan `time.Sleep(3s)`) — pastikan
   proses menunggu job selesai (dalam batas timeout `Stop`) sebelum exit, bukan langsung
   dibunuh di tengah jalan.
