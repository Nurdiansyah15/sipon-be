# Plan: Akademik Portal Santri (Non-Admin)

## Context

Sistem akademik sudah lengkap untuk admin (`/admin/akademik`). Namun santri belum bisa
mengakses informasi akademik mereka sendiri. Plan ini menambah portal santri (non-admin)
untuk:

1. **Summary** — periode aktif, status herregistrasi, tombol ajukan herreg jika belum.
2. **Program** — program yang diikuti santri.
3. **Kegiatan & Jadwal** — aktivitas wajib santri di periode aktif + jadwalnya.

**Scope tidak termasuk**: absensi, sesi, penilaian.

---

## Arsitektur & Keputusan Kunci

1. **Follow existing pattern** — sama seperti keuangan, kesantrian, psb: route group
   dengan `jwtAuth + principalLoad` saja, **tanpa** `RequirePermission`.
2. **Resolve user_id → santri_id** — gunakan `KesantrianReader.GetSantriByUserID`
   (perlu tambah method ini di port akademik).
3. **Prefix `/my/`** — semua endpoint santri pakai prefix `my` untuk clarity
   (mirip pattern `/keuangan/summary`, `/feedbacks/my`).
4. **Scope data ke santri** — setiap query otomatis filter berdasarkan santri_id
   yang di-resolve dari JWT. Santri tidak bisa query data santri lain.
5. **Herregistrasi self-service** — santri bisa **mengajukan** (status: pending),
   admin yang **complete/cancel** (flow existing tetap dipakai).
6. **Single active period** — semua data scoped ke periode yang sedang `open`.
   Jika ada multiple open periods, tampilkan yang paling baru (start_date tertinggi).
7. **Aktivitas wajib** = semua `ActivityPeriod` yang:
   - `academic_period_id` = periode aktif santri
   - `status` = `active`
   - `ActivityPeriodProgram` mencakup program santri, **atau** tidak ada scope
     (berarti berlaku untuk semua program)

---

## Backend: Endpoint yang Ditambahkan

### Port & Gateway Changes

**File**: `internal/modules/akademik/application/ports/kesantrian_reader.go`

```go
type KesantrianReader interface {
    GetSantriByID(ctx context.Context, santriID string) (*SantriBasicInfo, error)
    GetSantriByUserID(ctx context.Context, userID string) (*SantriBasicInfo, error) // TAMBAH
    ListActiveSantriWithUserID(ctx context.Context) ([]SantriBasicInfo, error)
}
```

**File**: `internal/modules/akademik/infrastructure/kesantriangateway/gateway.go`
Tambah method `GetSantriByUserID` yang delegasi ke `kesantrian.Contract.GetSantriByUserID`.

**File**: `internal/modules/akademik/module.go`
Wire gateway method baru.

---

### Router

**File**: `internal/modules/akademik/interfaces/http/router.go`

Tambah di group `akademik` (yang sudah ada, **sebelum** admin routes):

```go
// Santri-facing routes (no permission check)
akademik.GET("/my/summary", h.MySummary)
akademik.GET("/my/program", h.MyProgram)
akademik.GET("/my/kegiatan", h.MyActivities)
akademik.GET("/my/jadwal", h.MySchedules)
akademik.POST("/my/herregistrasi", h.ApplyHerregistrasi)
```

---

### Endpoint Details

#### GET /my/summary

Returns overview untuk dashboard santri.

```json
{
  "status": "success",
  "data": {
    "academic_period": {
      "id": "uuid",
      "code": "2026/2027-P1",
      "name": "Periode 1 2026/2027",
      "start_date": "2026-07-01",
      "end_date": "2026-12-31",
      "status": "open"
    },
    "herregistrasi": {
      "status": "completed",
      "registration_id": "uuid",
      "registered_at": "2026-07-15T10:00:00Z"
    },
    "program": {
      "id": "uuid",
      "code": "TAHFIDZ",
      "name": "Tahfidz"
    }
  }
}
```

**`herregistrasi.status`** bisa:
- `null` (belum ada record) → tampilkan tombol "Ajukan Herregistrasi"
- `"pending"` → tampilkan "Menunggu konfirmasi admin"
- `"completed"` → tampilkan badge "Sudah herregistrasi"
- `"cancelled"` → tampilkan "Dibatalkan" (bisa ajukan ulang?)

**Logic**:
1. Resolve user_id → santri_id via `GetSantriByUserID`
2. Cari periode aktif: `academic_periods WHERE status = 'open' ORDER BY start_date DESC LIMIT 1`
3. Cari herregistrasi santri di periode tsb: `santri_registrations WHERE santri_id = ? AND academic_period_id = ?`
4. Cari program santri: `santri_programs WHERE santri_id = ? AND is_active = true` → join `programs`

---

#### POST /my/herregistrasi

Santri mengajukan herregistrasi di periode aktif.

**Request**: tidak butuh body (otomatis pakai periode aktif).

**Response**: 201 + `SantriRegistrationResponse`

**Logic**:
1. Resolve user_id → santri_id
2. Cari periode aktif (`open`)
3. Validasi: belum ada herregistrasi di periode tsb (unique constraint)
4. Create `SantriRegistration` dengan status `pending`
5. Return response

**Error**:
- 409: sudah ada herregistrasi di periode aktif
- 404: tidak ada periode aktif (status `open`)
- 422: santri tidak punya program aktif

---

#### GET /my/program

Returns program aktif santri.

```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "code": "TAHFIDZ",
    "name": "Tahfidz",
    "status": "active",
    "started_at": "2026-07-01T00:00:00Z"
  }
}
```

**Error**:
- 404: santri belum punya program aktif

---

#### GET /my/kegiatan

Returns list aktivitas wajib santri di periode aktif.

```json
{
  "status": "success",
  "data": [
    {
      "id": "uuid",
      "activity_id": "uuid",
      "activity_code": "SHALAT_SUBUH",
      "activity_name": "Shalat Subuh Berjamaah",
      "activity_period_id": "uuid",
      "status": "active",
      "schedule_count": 1
    },
    {
      "id": "uuid",
      "activity_id": "uuid",
      "activity_code": "KAJIAN",
      "activity_name": "Kajian Kitab",
      "activity_period_id": "uuid",
      "status": "active",
      "schedule_count": 2
    }
  ]
}
```

**Logic**:
1. Resolve user_id → santri_id
2. Get periode aktif
3. Get program santri
4. Query `activity_periods`:
   ```sql
   SELECT ap.*, a.code as activity_code, a.name as activity_name
   FROM activity_periods ap
   JOIN activities a ON a.id = ap.activity_id
   WHERE ap.academic_period_id = $periode_id
     AND ap.status = 'active'
     AND (
       NOT EXISTS (SELECT 1 FROM activity_period_programs app WHERE app.activity_period_id = ap.id)
       OR EXISTS (SELECT 1 FROM activity_period_programs app
                  WHERE app.activity_period_id = ap.id AND app.program_id = $program_id)
     )
   ```
5. Count schedules per activity_period

**Error**:
- 404: tidak ada periode aktif
- 422: santri belum punya program

---

#### GET /my/jadwal

Returns jadwal untuk semua kegiatan wajib santri.

```json
{
  "status": "success",
  "data": [
    {
      "id": "uuid",
      "activity_period_id": "uuid",
      "activity_name": "Shalat Subuh Berjamaah",
      "activity_code": "SHALAT_SUBUH",
      "type": "daily",
      "start_time": "04:30:00",
      "end_time": "05:00:00"
    },
    {
      "id": "uuid",
      "activity_period_id": "uuid",
      "activity_name": "Kajian Kitab",
      "activity_code": "KAJIAN",
      "type": "weekly",
      "start_time": "19:30:00",
      "end_time": "21:00:00",
      "weekly_days": ["monday", "thursday"]
    }
  ]
}
```

**Logic**:
1. Resolve user_id → santri_id
2. Get periode aktif, program santri
3. Query activity_periods yang applicable (same filter as kegiatan)
4. Query schedules untuk activity_periods tersebut:
   ```sql
   SELECT s.*, a.name as activity_name, a.code as activity_code
   FROM activity_schedules s
   JOIN activity_periods ap ON ap.id = s.activity_period_id
   JOIN activities a ON a.id = ap.activity_id
   WHERE ap.academic_period_id = $periode_id
     AND ap.status = 'active'
     AND (scope program filter sama seperti kegiatan)
   ORDER BY
     CASE s.type WHEN 'daily' THEN 1 WHEN 'weekly' THEN 2 WHEN 'monthly' THEN 3 WHEN 'yearly' THEN 4 WHEN 'once' THEN 5 END,
     s.start_time
   ```
5. Include recurrence rules (weekly_days, monthly_days, yearly_dates)

**Error**:
- 404: tidak ada periode aktif

---

### DTOs Baru

**File**: `internal/modules/akademik/application/dto/santri_portal_dto.go`

```go
type MySummaryResponse struct {
    AcademicPeriod *AcademicPeriodResponse `json:"academic_period"`
    Herregistrasi  *HerregistrasiStatus    `json:"herregistrasi"`
    Program        *ProgramInfo            `json:"program"`
}

type HerregistrasiStatus struct {
    Status         string     `json:"status"`           // "none" | "pending" | "completed" | "cancelled"
    RegistrationID *string    `json:"registration_id"`
    RegisteredAt   *time.Time `json:"registered_at"`
}

type ProgramInfo struct {
    ID        string     `json:"id"`
    Code      string     `json:"code"`
    Name      string     `json:"name"`
    StartedAt *time.Time `json:"started_at"`
}

type MyActivityResponse struct {
    ID               string `json:"id"`
    ActivityID       string `json:"activity_id"`
    ActivityCode     string `json:"activity_code"`
    ActivityName     string `json:"activity_name"`
    ActivityPeriodID string `json:"activity_period_id"`
    Status           string `json:"status"`
    ScheduleCount    int    `json:"schedule_count"`
}

type MyScheduleResponse struct {
    ID               string   `json:"id"`
    ActivityPeriodID string   `json:"activity_period_id"`
    ActivityName     string   `json:"activity_name"`
    ActivityCode     string   `json:"activity_code"`
    Type             string   `json:"type"`
    StartDate        *string  `json:"start_date"`
    EndDate          *string  `json:"end_date"`
    StartTime        string   `json:"start_time"`
    EndTime          string   `json:"end_time"`
    WeeklyDays       []string `json:"weekly_days,omitempty"`
    MonthlyDays      []int    `json:"monthly_days,omitempty"`
    YearlyDates      []YearlyDateDTO `json:"yearly_dates,omitempty"`
}
```

---

### Queries Baru

**File**: `internal/modules/akademik/application/query/`

| File | Use Case | Deskripsi |
|------|----------|-----------|
| `get_active_period.go` | `GetActivePeriodUseCase` | Cari periode open terbaru |
| `get_my_summary.go` | `GetMySummaryUseCase` | Aggregate: period + herreg + program |
| `get_my_program.go` | `GetMyProgramUseCase` | Get santri's active program |
| `list_my_activities.go` | `ListMyActivitiesUseCase` | List applicable activities |
| `list_my_schedules.go` | `ListMySchedulesUseCase` | List applicable schedules |

### Commands Baru

**File**: `internal/modules/akademik/application/command/apply_herregistrasi.go`

Use case: `ApplyHerregistrasiUseCase`
- Resolve santri dari user_id
- Validate periode aktif ada
- Validate belum ada herregistrasi di periode tsb
- Create `SantriRegistration` status `pending`

---

### Repository Changes

**File**: `internal/modules/akademik/domain/academic_period/repository/interfaces.go`
Tambah: `FindOpen(ctx context.Context) (*entity.AcademicPeriod, error)`
— return single open period dengan start_date tertinggi.

**File**: `internal/modules/akademik/domain/activity_period/repository/interfaces.go`
Tambah: `ListByPeriodAndProgram(ctx, periodID, programID string) ([]*entity.ActivityPeriod, error)`
— return activity_periods yang applicable untuk program tertentu (include yang tanpa scope).

**File**: `internal/modules/akademik/domain/activity_schedule/repository/interfaces.go`
Tambah: `ListByActivityPeriodIDs(ctx, periodIDs []string) ([]*entity.ActivitySchedule, error)`
— return schedules untuk multiple activity_periods.

---

## Frontend: Struktur Halaman

### Routing

**Prefix**: `/akademik` (di luar `/admin`)

```
pages/
  akademik/
    index.vue                    # Dashboard summary
    kegiatan.vue                 # List kegiatan wajib
    jadwal.vue                   # List jadwal
```

### Layout

Gunakan layout santri existing (same as `/keuangan`, `/kesantrian`).
Tambahkan link di sidebar santri navbar.

### Store

**File**: `app/stores/akademik-santri.ts` (baru, terpisah dari admin store)

```typescript
// State
activePeriod: AcademicPeriod | null
herregistrasiStatus: { status: string, registration_id?: string }
program: ProgramInfo | null
activities: MyActivity[]
schedules: MySchedule[]
isLoading: boolean
error: string | null

// Actions
fetchSummary()
applyHerregistrasi()
fetchActivities()
fetchSchedules()
```

### Types

**File**: `shared/types/AkademikSantri.ts` (baru)

```typescript
export interface MySummary {
  academic_period: AcademicPeriod | null
  herregistrasi: {
    status: 'none' | 'pending' | 'completed' | 'cancelled'
    registration_id?: string
    registered_at?: string
  }
  program: {
    id: string
    code: string
    name: string
    started_at: string
  } | null
}

export interface MyActivity {
  id: string
  activity_id: string
  activity_code: string
  activity_name: string
  activity_period_id: string
  status: string
  schedule_count: number
}

export interface MySchedule {
  id: string
  activity_period_id: string
  activity_name: string
  activity_code: string
  type: ActivityScheduleType
  start_date?: string
  end_date?: string
  start_time: string
  end_time: string
  weekly_days?: DayOfWeek[]
  monthly_days?: number[]
  yearly_dates?: YearlyDate[]
}
```

---

### Halaman: Dashboard (`/akademik`)

```
┌──────────────────────────────────────────────────┐
│  Akademik                                        │
├──────────────────────────────────────────────────┤
│                                                  │
│  ┌─────────────┐  ┌─────────────┐               │
│  │ Periode     │  │ Herregistrasi│               │
│  │ Aktif       │  │             │               │
│  │             │  │  ✓ Sudah    │  (atau)       │
│  │ 2026/2027   │  │  ⏳ Menunggu│  (atau)       │
│  │ -P1         │  │  📋 Ajukan  │               │
│  │             │  │             │               │
│  │ 1 Jul -     │  └─────────────┘               │
│  │ 31 Des 2026 │                                 │
│  └─────────────┘  ┌─────────────┐               │
│                    │ Program     │               │
│                    │             │               │
│                    │ 📖 TAHFIDZ  │               │
│                    │ Tahfidz     │               │
│                    └─────────────┘               │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ Kegiatan Wajib (3)                [→]    │    │
│  │                                          │    │
│  │  • Shalat Subuh Berjamaah    (1 jadwal)  │    │
│  │  • Kajian Kitab              (2 jadwal)  │    │
│  │  • Mutaba'ah                (1 jadwal)   │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ Jadwal Hari Ini                    [→]   │    │
│  │                                          │    │
│  │  04:30 - 05:00  Shalat Subuh            │    │
│  │  19:30 - 21:00  Kajian Kitab (Sen,Kam)  │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Cards**:
1. **Periode Aktif** — nama periode + tanggal range + status badge
2. **Herregistrasi** — status badge + action button jika belum
3. **Program** — program name + code badge
4. **Kegiatan Wajib** — list kegiatan (max 5, show "Lihat semua" link)
5. **Jadwal Hari Ini** — list jadwal hari ini (filtered by current day)

**Herregistrasi flow**:
- `none` → tombol "Ajukan Herregistrasi" → POST /my/herregistrasi → reload summary
- `pending` → badge amber "Menunggu konfirmasi"
- `completed` → badge green "Sudah herregistrasi"
- `cancelled` → badge red + opsi ajukan ulang? (tombol "Ajukan Herregistrasi")

---

### Halaman: Kegiatan (`/akademik/kegiatan`)

```
┌──────────────────────────────────────────────────┐
│  ← Kembali ke Akademik                           │
├──────────────────────────────────────────────────┤
│  Kegiatan Wajib — Periode 1 2026/2027           │
├──────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ Shalat Subuh Berjamaah                   │    │
│  │ SHALAT_SUBUH                             │    │
│  │                                          │    │
│  │ Jadwal: 1                                │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ Kajian Kitab                             │    │
│  │ KAJIAN                                   │    │
│  │                                          │    │
│  │ Jadwal: 2                                │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ Mutaba'ah                                │    │
│  │ MUTABA'AH                                │    │
│  │                                          │    │
│  │ Jadwal: 1                                │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
└──────────────────────────────────────────────────┘
```

List kegiatan dengan info jumlah jadwal. Klik kegiatan → expand jadwal atau navigasi ke `/akademik/jadwal?activity_period_id=xxx`.

---

### Halaman: Jadwal (`/akademik/jadwal`)

```
┌──────────────────────────────────────────────────┐
│  ← Kembali ke Akademik                           │
├──────────────────────────────────────────────────┤
│  Jadwal Kegiatan — Periode 1 2026/2027          │
├──────────────────────────────────────────────────┤
│                                                  │
│  Filter: [Semua] [Harian] [Mingguan] [Bulanan]  │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ 📅 Shalat Subuh Berjamaah                │    │
│  │                                          │    │
│  │ Tipe: Harian                             │    │
│  │ Waktu: 04:30 - 05:00                     │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ 📅 Kajian Kitab                          │    │
│  │                                          │    │
│  │ Tipe: Mingguan                           │    │
│  │ Hari: Senin, Kamis                       │    │
│  │ Waktu: 19:30 - 21:00                     │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ 📅 Mutaba'ah                             │    │
│  │                                          │    │
│  │ Tipe: Harian                             │    │
│  │ Waktu: 20:00 - 20:30                     │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Display per schedule**:
- Activity name (dari activity_code + activity_name)
- Type badge (Harian/Mingguan/Bulanan/Tahunan/Sekali)
- Waktu (start_time - end_time)
- Recurrence detail:
  - `weekly`: hari-hari (Senin, Kamis)
  - `monthly`: tanggal (5, 20)
  - `yearly`: bulan-tanggal (17 Agustus)
  - `once`: tanggal spesifik (start_date)
- Date range jika ada (start_date - end_date)

**Filter**: dropdown/tabs untuk filter by type.

---

## Data Flow

```
Santri Login (JWT)
    │
    ▼
GET /my/summary
    │
    ├── resolve user_id → santri_id (via kesantrian)
    ├── get active period (status=open)
    ├── get santri's program (santri_programs WHERE is_active=true)
    ├── get herregistrasi status (santri_registrations)
    │
    ▼
Response: { period, herregistrasi, program }
    │
    ▼
GET /my/kegiatan
    │
    ├── resolve user_id → santri_id
    ├── get active period
    ├── get santri's program
    ├── query activity_periods (filtered by period + program scope)
    │
    ▼
Response: [{ activity, schedule_count }, ...]
    │
    ▼
GET /my/jadwal
    │
    ├── resolve user_id → santri_id
    ├── get active period
    ├── get santri's program
    ├── query activity_periods (filtered)
    ├── query activity_schedules (for those periods)
    ├── include recurrence rules
    │
    ▼
Response: [{ schedule, activity_name, type, times, recurrence }, ...]
```

---

## Error Handling

| Scenario | HTTP Status | Message |
|----------|-------------|---------|
| User tidak punya santri record | 404 | "Profil santri tidak ditemukan" |
| Tidak ada periode aktif | 404 | "Tidak ada periode akademik aktif" |
| Santri belum punya program | 422 | "Anda belum terdaftar di program manapun" |
| Ajukan herregistrasi tapi sudah ada | 409 | "Anda sudah mengajukan herregistrasi di periode ini" |
| Ajukan herregistrasi, periode belum open | 404 | "Periode akademik belum dibuka" |

---

## Fase Pengerjaan

### Fase 1 — Backend: Port & Gateway
- [ ] Tambah `GetSantriByUserID` di `KesantrianReader` port
- [ ] Implementasi di gateway adapter
- [ ] Wire di `module.go`
- [ ] Test: `go build ./...`

### Fase 2 — Backend: Repository Methods
- [ ] `FindOpen` di `academic_period` repo
- [ ] `ListByPeriodAndProgram` di `activity_period` repo (dengan scope filter)
- [ ] `ListByActivityPeriodIDs` di `activity_schedule` repo
- [ ] Migration: tambahkan index jika perlu untuk performance
- [ ] Test: `go build ./...`

### Fase 3 — Backend: Queries & Commands
- [ ] `GetActivePeriodUseCase`
- [ ] `GetMySummaryUseCase`
- [ ] `GetMyProgramUseCase`
- [ ] `ListMyActivitiesUseCase`
- [ ] `ListMySchedulesUseCase`
- [ ] `ApplyHerregistrasiUseCase`
- [ ] DTOs baru di `santri_portal_dto.go`
- [ ] Test: `go build ./...`

### Fase 4 — Backend: Handler & Router
- [ ] Tambah handler methods di `handler.go`
- [ ] Register routes di `router.go`
- [ ] Smoke test dengan curl/Postman
- [ ] Test: `go build ./...`

### Fase 5 — Frontend: Types & Store
- [ ] Buat `shared/types/AkademikSantri.ts`
- [ ] Buat `app/stores/akademik-santri.ts`
- [ ] Test: type-check lolos

### Fase 6 — Frontend: Halaman Dashboard
- [ ] Buat `pages/akademik/index.vue`
- [ ] Summary cards (period, herregistrasi, program)
- [ ] Quick list kegiatan
- [ ] Quick list jadwal hari ini
- [ ] Herregistrasi apply flow (button → POST → reload)
- [ ] Test: render bersih, data tampil

### Fase 7 — Frontend: Halaman Kegiatan & Jadwal
- [ ] Buat `pages/akademik/kegiatan.vue`
- [ ] Buat `pages/akademik/jadwal.vue`
- [ ] Filter jadwal by type
- [ ] Recurrence display (weekly days, monthly dates, etc)
- [ ] Test: render bersih, filter bekerja

### Fase 8 — Frontend: Navigation & Polish
- [ ] Tambah link di sidebar santri navbar
- [ ] Empty states (belum ada periode, belum ada kegiatan)
- [ ] Loading states
- [ ] Error handling (toast messages)
- [ ] Mobile responsive
- [ ] Test: end-to-end flow

---

## Verifikasi

1. `go build ./...` lolos di tiap fase backend.
2. `npx vue-tsc --noEmit` lolos di tiap fase frontend.
3. Smoke test E2E:
   - Login sebagai santri (user yang punya santri record)
   - Dashboard summary tampil: periode aktif, status herreg, program
   - Jika belum herreg, tombol "Ajukan" berfungsi
   - Klik kegiatan → list kegiatan muncul dengan schedule count
   - Klik jadwal → list jadwal muncul dengan recurrence info
   - Filter jadwal by type bekerja
4. Edge cases:
   - Santri tanpa program → error message jelas
   - Tidak ada periode aktif → empty state
   - Periode closed/archived → tidak muncul di summary
   - Santri sudah herreg completed → tombol hilang, badge hijau
5. Security:
   - Santri A tidak bisa akses data Santri B
   - Endpoint hanya return data scoped ke santri yang login
   - Ajukan herregistrasi untuk periode lain → error
