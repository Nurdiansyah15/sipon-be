# Plan: Integrasi Absensi dengan Mesin Fingerprint (+ Sandbox)

## Context

Saat ini absensi (`akademik` module) dicatat dengan dua cara: admin input manual per-santri
(`RecordAttendanceUseCase`), atau santri input NIS sendiri di halaman presensi
(`CheckinByNISUseCase`, lihat `internal/modules/akademik/application/command/checkin_by_nis.go`).
Keduanya butuh aksi manual dari orang.

Kebutuhan baru: mesin fingerprint mengirim data scan ke sebuah database dengan skema
`sn, scan_date, pin, verifymode, inoutmode, deviceip` — `pin` berisi NIS santri, `scan_date`
adalah waktu scan. Sistem butuh:

1. **"Get scan info"** — proses yang mengambil scan **non-duplikat per pin** dalam rentang
   waktu yang ditentukan oleh jadwal sesi (`ActivitySession.StartsAt`..`EndsAt`), lalu
   mencatatnya sebagai kehadiran — menggantikan/melengkapi input manual NIS.
2. **Sandbox** — generator "mesin fingerprint palsu" yang menulis data scan dengan skema
   yang sama ke database kita, supaya alur di atas bisa dikembangkan & dites tanpa hardware
   fisik.

Keputusan yang sudah dikonfirmasi user:
- **Sumber data**: untuk sekarang tabel scan log dibuat **di database Postgres sipon sendiri**
  (bukan koneksi ke DB eksternal terpisah). Codebase belum punya pola multi-datasource
  (single `*sql.DB` dipakai semua module, lihat `cmd/api/main.go:45-51`) dan belum ada driver
  DB lain di `go.mod` selain `pgx`. Interface repository dirancang generik supaya kalau nanti
  mesin asli menulis ke DB eksternal (mis. MySQL), cukup ganti implementasi
  `ScanLogRepository` tanpa mengubah use case/kontrak di atasnya.
- **Trigger sync**: manual dulu — admin klik tombol di halaman sesi, tidak ada background
  poller/cron (codebase belum punya infrastruktur worker/scheduler, lihat tidak ada
  `cron`/`ticker` di manapun kecuali goroutine start-server).

## Arsitektur

Modul baru `internal/modules/fingerprint`, mandiri (tidak depend ke module lain), mengikuti
pola modular-monolith yang sudah ada (domain/application/infrastructure/interfaces + `module.go`
+ `contract.go`, lihat contoh `internal/modules/akademik/`). `akademik` mengonsumsi modul ini
lewat `Contract` + port + gateway, sama seperti pola `kesantrian` → `kesantriangateway`
(`internal/modules/akademik/infrastructure/kesantriangateway/gateway.go`).

```
fingerprint module                          akademik module
─────────────────                           ────────────────
domain/scanlog/entity/scan_log.go           application/ports/fingerprint_reader.go (BARU)
domain/scanlog/repository/interfaces.go     infrastructure/fingerprintgateway/gateway.go (BARU)
infrastructure/persistence/                 application/command/
  postgres_scan_log_repo.go                   sync_attendance_from_fingerprint.go (BARU)
application/command/simulate_scan.go        application/dto/sync_fingerprint_dto.go (BARU)
  (sandbox: insert 1 scan palsu)            interfaces/http/handler.go (UPDATE: +SyncFingerprint)
application/query/list_scan_logs.go         interfaces/http/router.go (UPDATE: +route)
  (get scan info: dedup per pin, range)     module.go (UPDATE: wire fingerprint dependency)
application/query/list_distinct_pins.go
interfaces/http/handler.go
interfaces/http/router.go
module.go, contract.go
```

Tabel baru (migrasi Postgres), skema identik dengan yang dikirim mesin fingerprint:

```sql
-- migrations/20260812120000_create_fingerprint_scan_logs.up.sql
CREATE TABLE IF NOT EXISTS fingerprint_scan_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sn          VARCHAR(50) NOT NULL,
    scan_date   TIMESTAMPTZ NOT NULL,
    pin         VARCHAR(50) NOT NULL,
    verifymode  INT,
    inoutmode   INT,
    deviceip    VARCHAR(45),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_fingerprint_scan_logs_pin_scandate ON fingerprint_scan_logs(pin, scan_date);
CREATE INDEX idx_fingerprint_scan_logs_scandate ON fingerprint_scan_logs(scan_date);
```
(+ file `.down.sql` yang `DROP TABLE`, ikuti pola pasangan file existing di `migrations/`.)

Catatan: `pin` disimpan sebagai teks (bukan asumsi numerik) karena mesin fingerprint
umumnya menyimpan PIN sebagai string; NIS santri (`internal/modules/kesantrian/domain/santri/valueobject/nis.go`)
berformat 10 digit (`1000` + 1 digit gender + 5 digit) — cocok dipakai sebagai PIN pada mesin.

## Detail: Modul `fingerprint`

**Entity** — `domain/scanlog/entity/scan_log.go`:
```go
type ScanLog struct {
    ID, SN, PIN, DeviceIP string
    ScanDate              time.Time
    VerifyMode, InOutMode int
    CreatedAt             time.Time
}
func NewScanLog(id, sn, pin, deviceIP string, scanDate time.Time, verifyMode, inOutMode int) (*ScanLog, error)
```
Validasi minimal: `id`, `sn`, `pin` tidak boleh kosong.

**Repository** — `domain/scanlog/repository/interfaces.go`:
```go
type ScanPin struct {
    PIN         string
    SN          string
    FirstScanAt time.Time
}

type ScanLogRepository interface {
    Insert(ctx context.Context, log *entity.ScanLog) error
    // ListDistinctPinInRange returns the first scan per pin within [from, to],
    // dedup dilakukan di query (SELECT DISTINCT ON (pin) ... ORDER BY pin, scan_date).
    ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]ScanPin, error)
    ListInRange(ctx context.Context, from, to time.Time) ([]*entity.ScanLog, error) // untuk debug/inspeksi
}
```

**Persistence** — `infrastructure/persistence/postgres_scan_log_repo.go`: implementasi biasa
`database/sql` + `pgx`, sama gaya dengan `postgres_attendance_repo.go`. Query dedup:
```sql
SELECT DISTINCT ON (pin) pin, sn, scan_date
FROM fingerprint_scan_logs
WHERE scan_date >= $1 AND scan_date <= $2
ORDER BY pin, scan_date ASC
```

**Application (sandbox)** — `application/command/simulate_scan.go`:
- `SimulateScanUseCase.Execute(ctx, input)`: `input{SN *string, PIN string (required), ScanDate *time.Time, VerifyMode, InOutMode *int, DeviceIP *string}`.
  Default `SN` = `"SANDBOX-DEVICE-01"`, `ScanDate` = `time.Now()` jika kosong. Insert 1 baris.
- (opsional, kalau berguna untuk testing) `SimulateBulkScanUseCase`: terima list PIN + jumlah
  scan acak dalam rentang waktu, untuk generate data uji dengan cepat.

**Application (get scan info)** — `application/query/list_distinct_pins.go`:
- `ListDistinctPinsUseCase.Execute(ctx, from, to time.Time) ([]dto.ScanPin, error)` — tipis,
  cuma memanggil `repo.ListDistinctPinInRange`. Dipakai oleh HTTP handler (debug) dan diekspos
  lewat `Contract` untuk dikonsumsi module lain.

**Contract** — `contract.go`:
```go
type ScanPin struct {
    PIN         string
    FirstScanAt time.Time
}
type Contract interface {
    ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]ScanPin, error)
}
```

**HTTP** — `interfaces/http/router.go`, base path `/api/v1/web/fingerprint`:
| Method | Path | Guard | Deskripsi |
|---|---|---|---|
| POST | `/sandbox/scan` | JWT + `manage_akademik`, **dan** hanya didaftarkan jika `cfg.Fingerprint.SandboxEnabled` | Simulasi 1 scan palsu |
| GET | `/scans` | JWT + `manage_akademik` | List scan mentah dalam rentang (debug), query `from`, `to` |

Route sandbox didaftarkan secara kondisional di `RegisterRoutes` (`if sandboxEnabled { ... }`)
supaya di production (`FINGERPRINT_SANDBOX_ENABLED` default `false`) endpoint ini tidak ada
sama sekali — bukan cuma disembunyikan di permission.

**Config** — `internal/shared/config/config.go`: tambah
```go
type FingerprintConfig struct {
    SandboxEnabled bool
}
// di Config struct: Fingerprint FingerprintConfig
// di Load(): Fingerprint: FingerprintConfig{SandboxEnabled: getEnv("FINGERPRINT_SANDBOX_ENABLED", "false") == "true"}
```

**Wiring** — `module.go` ikut pola `akademik/module.go` (`NewModule(db *sql.DB, cfg *config.Config, jwtAuth, principalLoad gin.HandlerFunc) *Module`), didaftarkan di `cmd/api/main.go` sebelum `akademik`:
```go
fingerprint := fingerprintModule.NewModule(db, cfg, identity.AuthMiddleware(), identity.PrincipalMiddleware())
akademik := akademikModule.NewModule(db, cfg, kesantrian, fingerprint, identity.AuthMiddleware(), identity.PrincipalMiddleware())
...
fingerprint.RegisterRoutes(engine)
akademik.RegisterRoutes(engine)
```

## Detail: Update Modul `akademik`

**Port** — `application/ports/fingerprint_reader.go` (BARU), mirror `kesantrian_reader.go`:
```go
type FingerprintScanPin struct {
    PIN         string
    FirstScanAt time.Time
}
type FingerprintReader interface {
    ListDistinctPinInRange(ctx context.Context, from, to time.Time) ([]FingerprintScanPin, error)
}
```

**Gateway** — `infrastructure/fingerprintgateway/gateway.go` (BARU), delegasi ke
`fingerprint.Contract`, sama gaya `kesantriangateway/gateway.go`.

**Use case baru** — `application/command/sync_attendance_from_fingerprint.go`:
```go
type SyncAttendanceFromFingerprintUseCase struct {
    sessionRepo       sesRepo.ActivitySessionRepository
    fingerprintReader ports.FingerprintReader
    checkin           *CheckinByNISUseCase // REUSE langsung, tidak ada logic baru untuk validasi
}

func (uc *SyncAttendanceFromFingerprintUseCase) Execute(ctx context.Context, sessionID string) (*dto.SyncFingerprintResponse, error) {
    session, err := uc.sessionRepo.FindByID(ctx, sessionID)
    if err != nil { return nil, 404 }
    if session.Status != "open" { return nil, 422 "sesi tidak terbuka" }

    to := time.Now()
    if session.EndsAt.Before(to) { to = session.EndsAt } // jangan ambil scan setelah sesi selesai

    scans, err := uc.fingerprintReader.ListDistinctPinInRange(ctx, session.StartsAt, to)
    if err != nil { return nil, 500 }

    resp := &dto.SyncFingerprintResponse{TotalScans: len(scans)}
    for _, scan := range scans {
        _, err := uc.checkin.Execute(ctx, sessionID, scan.PIN)
        switch {
        case err == nil:
            resp.Recorded++
        case isConflict(err): // NIS sudah tercatat -> bukan error, cuma skip
            resp.Skipped++
        default:
            resp.Errors = append(resp.Errors, dto.SyncFingerprintError{PIN: scan.PIN, Reason: err.Error()})
        }
    }
    return resp, nil
}
```
`CheckinByNISUseCase.Execute` (sudah ada, **tidak diubah**) sudah menjalankan semua validasi
yang relevan: sesi open, santri aktif, herreg completed, program membership, dan cek duplikat
(`internal/modules/akademik/application/command/checkin_by_nis.go:54-110`) — jadi scan dari
mesin fingerprint melalui jalur eligibility yang **sama persis** dengan check-in manual by NIS,
tanpa duplikasi logic.

Untuk membedakan "sudah tercatat" (skip, bukan error) dari error lain, pakai pola
`errors.As` yang sudah dipakai di codebase (lihat
`internal/modules/keuangan/application/command/verify_payment.go:153-171`):
```go
func isConflict(err error) bool {
    var ke *kernel.AppError
    return errors.As(err, &ke) && ke.Code == application.ErrCodeConflict
}
```

**DTO** — `application/dto/sync_fingerprint_dto.go` (BARU):
```go
type SyncFingerprintResponse struct {
    TotalScans int                    `json:"total_scans"`
    Recorded   int                    `json:"recorded"`
    Skipped    int                    `json:"skipped"`
    Errors     []SyncFingerprintError `json:"errors"`
}
type SyncFingerprintError struct {
    PIN    string `json:"pin"`
    Reason string `json:"reason"`
}
```

**Handler + Router**: tambah `h.SyncFingerprint` di `interfaces/http/handler.go`, route baru
di grup admin sessions yang sudah ada (`interfaces/http/router.go:129-142`):
```go
sessions.POST("/:id/sync-fingerprint", h.SyncFingerprint) // JWT + manage_akademik, sama seperti route sekitarnya
```

**Wiring** — `module.go`: `NewModule` akademik menerima param baru `fingerprintContract fingerprint.Contract`, buat `fingerprintGW := fingerprintgateway.New(fingerprintContract)`, lalu:
```go
syncFingerprintUC := command.NewSyncAttendanceFromFingerprintUseCase(sessionRepo, fingerprintGW, checkinUC)
```
tambahkan ke `NewAkademikHandler(...)` dan ke `main.go` (`akademikModule.NewModule(db, cfg, kesantrian, fingerprint, ...)`).

## Fase Pengerjaan

1. **Modul `fingerprint` — domain & persistence**: migrasi tabel, entity `ScanLog`,
   `ScanLogRepository` + implementasi Postgres. `go build ./...`.
2. **Modul `fingerprint` — sandbox & get-scan-info**: `SimulateScanUseCase`,
   `ListDistinctPinsUseCase`, handler + router (endpoint sandbox dikondisikan
   `cfg.Fingerprint.SandboxEnabled`), `contract.go`, `module.go`. Tambah `FingerprintConfig`
   di `config.go`. `go build ./...`.
3. **Wire modul baru ke `main.go`** (sebelum akademik, tanpa dependency lain). Jalankan
   migrasi, smoke test manual: `POST /sandbox/scan` beberapa kali dengan PIN berbeda →
   `GET /scans` menunjukkan data masuk.
4. **Akademik — konsumsi fingerprint**: port `FingerprintReader`, gateway
   `fingerprintgateway`, use case `SyncAttendanceFromFingerprintUseCase`, DTO, handler,
   route `/sessions/:id/sync-fingerprint`, update `module.go` + `main.go`. `go build ./...`.
5. **Testing**:
   - Unit test `SyncAttendanceFromFingerprintUseCase`: recorded/skipped/error dihitung benar,
     sesi tidak-open ditolak, scan di luar rentang [starts_at, min(ends_at, now)] tidak diambil.
   - Unit test dedup: dua scan dengan pin sama dalam rentang → cuma diproses sekali (dijamin
     query `DISTINCT ON`, tapi tetap divalidasi lewat repository test/integration test).
   - `go test ./...`.

## Verifikasi End-to-End

1. Jalankan migrasi baru, `go build ./...`, `go test ./...`.
2. Set `FINGERPRINT_SANDBOX_ENABLED=true` di `.env` lokal.
3. Buat academic period + activity + schedule + generate session (pakai endpoint admin yang
   sudah ada), buka sesi (`POST /sessions/:id/open`).
4. Simulasikan scan mesin: `POST /api/v1/web/fingerprint/sandbox/scan` dengan `pin` = NIS
   santri yang sudah herreg & terdaftar program terkait, `scan_date` di dalam rentang
   `starts_at..ends_at` sesi.
5. Panggil `POST /api/v1/web/akademik/sessions/:id/sync-fingerprint` → cek response
   `{recorded: 1, skipped: 0, errors: []}`, lalu cek `GET /sessions/:id/attendance` — santri
   tercatat `present`.
6. Ulangi sync tanpa scan baru → `recorded: 0, skipped: 1` (idempotent, tidak dobel).
7. Simulasikan scan dengan PIN yang tidak match NIS manapun / santri belum herreg → cek masuk
   ke `errors[]` dengan alasan yang sesuai, bukan bikin seluruh sync gagal.
