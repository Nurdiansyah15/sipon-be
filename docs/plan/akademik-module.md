# Plan: Module Akademik (Pondok Pesantren Academic & Activity Domain)

## Context

Sistem SIPON belum memiliki modul akademik untuk mengelola kegiatan pendidikan pesantren.
Domain utama yang perlu dibangun:

```
Program
    ↓
Santri
    ↓
Academic Period
    ↓
Herregistrasi (SantriRegistration)
    ↓
Activity
    ↓
Activity Schedule
    ↓
Activity Session
    ↓
Attendance
```

Sistem ini **bukan** sistem universitas. Tidak ada Course, KRS, KHS, Lecturer,
Course Enrollment. Domain pesantren: Program, Santri, AcademicPeriod, Activity, Session, Attendance.

Scope fase ini:
- Master Program (TAHFIDZ, KITAB)
- Master Santri (sudah ada di `kesantrian`, akademik hanya pakai referensi)
- Academic Period (periode akademik generik)
- Santri Registration / Herregistrasi (registrasi santri per periode)
- Activity (master kegiatan)
- Activity Period (aktivasi activity per periode)
- Activity Period Program (scope program per activity-period)
- Activity Schedule (pola recurrence)
- Activity Schedule Weekly / Monthly / Yearly (detail recurrence)
- Activity Session (occurrence konkret)
- Attendance (absensi santri)

Scope yang **TIDAK** dikerjakan:
- Penilaian / Assessment / Score / Grade
- Pelanggaran / Violation / Takziran / Penebusan
- Billing / Payment (detail)
- Kelulusan / Wisuda
- Kurikulum

## Keputusan Arsitektur Kunci

1. **Module baru `akademik`** — mengikuti pola modular monolith DDD yang sama
   seperti `kesantrian`, `psb`, `keuangan`. Struktur: domain/, application/,
   infrastructure/, interfaces/http/.
2. **Santri tetap di `kesantrian`** — akademik tidak menduplikasi data santri.
   Mengakses data santri lewat `kesantrian.Contract` (port+gateway pattern).
3. **Program adalah entitas terpisah** — bukan hard-code string. Tabel `programs`
   di module akademik. `santri.program` di kesantrian tetap ada sebagai data
   historis; akademik menggunakan `program_id` di tabel `santri_registrations`
   untuk scope kegiatan (tidak perlu FK ke kesantrian — cross-module).
4. **ActivityPeriod menjembatani Activity dan AcademicPeriod** — activity
   adalah master, activity_period menentukan apakah aktif pada periode tertentu.
5. **ActivitySchedule terpisah dari ActivitySession** — schedule adalah aturan,
   session adalah kejadian konkret. Schedule tidak boleh langsung menjadi session.
6. **Recurrence detail per tipe** — Weekly, Monthly, Yearly masing-masing tabel
   detail. Jangan simpan recurrence sebagai string.
7. **Tidak ada FK lintas module** — `santri_id` di attendance/santri_registration
   adalah plain UUID (cross-module ke kesantrian).
8. **Soft delete untuk histori** — entity yang sudah digunakan dalam transaksi
   akademik tidak di-hard-delete. Gunakan `deleted_at`.

## Domain Model

### Program
Master program pendidikan pesantren.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| code | VARCHAR(20) UNIQUE | e.g. TAHFIDZ, KITAB |
| name | VARCHAR(100) | Nama program |
| status | VARCHAR(20) | active / inactive |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | nullable |

Constants:
```
ProgramCode: TAHFIDZ, KITAB
ProgramStatus: active, inactive
```

### SantriRegistration (Herregistrasi)
Registrasi santri pada suatu academic period.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| santri_id | UUID NOT NULL | Cross-module ke kesantrian (no FK) |
| academic_period_id | UUID NOT NULL FK → academic_periods | |
| status | VARCHAR(20) | pending / completed / cancelled |
| registered_at | TIMESTAMPTZ | Waktu registrasi selesai |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | nullable |

Unique: `(santri_id, academic_period_id) WHERE deleted_at IS NULL`

Constants:
```
SantriRegistrationStatus: pending, completed, cancelled
```

### AcademicPeriod
Periode akademik pesantren (bukan semester — generik).

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| code | VARCHAR(50) UNIQUE | e.g. 2026/2027-P1 |
| name | VARCHAR(100) | e.g. Periode 1 2026/2027 |
| start_date | DATE NOT NULL | |
| end_date | DATE NOT NULL | |
| status | VARCHAR(20) | draft / open / closed / archived |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | nullable |

Constants:
```
AcademicPeriodStatus: draft, open, closed, archived
```

### Activity
Master kegiatan.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| code | VARCHAR(50) UNIQUE | e.g. SHALAT_SUBUH, KAJIAN |
| name | VARCHAR(200) | Nama kegiatan |
| status | VARCHAR(20) | active / inactive |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | nullable |

Status = status master. Bukan status per period.

Constants:
```
ActivityStatus: active, inactive
```

### ActivityPeriod
Aktivasi activity pada academic period tertentu.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| activity_id | UUID NOT NULL FK → activities | |
| academic_period_id | UUID NOT NULL FK → academic_periods | |
| status | VARCHAR(20) | active / inactive |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | nullable |

Unique: `(activity_id, academic_period_id) WHERE deleted_at IS NULL`

Constants:
```
ActivityPeriodStatus: active, inactive
```

### ActivityPeriodProgram
Scope program per activity-period. Menentukan activity berlaku untuk program apa.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| activity_period_id | UUID NOT NULL FK → activity_periods ON DELETE CASCADE | |
| program_id | UUID NOT NULL FK → programs | |

Unique: `(activity_period_id, program_id)`

Jika activity **tidak punya** ActivityPeriodProgram record → berlaku untuk
semua program (aturan default, didokumentasikan eksplisit).

### ActivitySchedule
Pola recurrence kegiatan.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| activity_period_id | UUID NOT NULL FK → activity_periods | |
| type | VARCHAR(20) | once / daily / weekly / monthly / yearly |
| start_date | DATE | Rentang mulai berlakunya schedule (opsional untuk ONCE) |
| end_date | DATE | Rentang akhir berlakunya schedule (opsional) |
| start_time | TIME | Waktu mulai |
| end_time | TIME | Waktu selesai |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | nullable |

Constants:
```
ActivityScheduleType: once, daily, weekly, monthly, yearly
DayOfWeek: monday, tuesday, wednesday, thursday, friday, saturday, sunday
```

### ActivityScheduleWeekly
Detail recurrence untuk schedule type = weekly.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| schedule_id | UUID NOT NULL FK → activity_schedules ON DELETE CASCADE | |
| day_of_week | VARCHAR(10) | e.g. monday, thursday |

Satu schedule weekly bisa punya multiple days (satu row per hari).

### ActivityScheduleMonthly
Detail recurrence untuk schedule type = monthly.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| schedule_id | UUID NOT NULL FK → activity_schedules ON DELETE CASCADE | |
| day_of_month | INT | 1-31 |

### ActivityScheduleYearly
Detail recurrence untuk schedule type = yearly.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| schedule_id | UUID NOT NULL FK → activity_schedules ON DELETE CASCADE | |
| month | INT | 1-12 |
| day | INT | 1-31 |

### ActivitySession
Occurrence konkret dari kegiatan.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| activity_schedule_id | UUID NOT NULL FK → activity_schedules | |
| starts_at | TIMESTAMPTZ NOT NULL | Waktu mulai session |
| ends_at | TIMESTAMPTZ NOT NULL | Waktu selesai session |
| status | VARCHAR(20) | scheduled / open / completed / cancelled |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | nullable |

Constants:
```
ActivitySessionStatus: scheduled, open, completed, cancelled
```

### Attendance
Absensi santri per session.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| activity_session_id | UUID NOT NULL FK → activity_sessions | |
| santri_id | UUID NOT NULL | Cross-module ke kesantrian (no FK) |
| status | VARCHAR(20) | present / absent / excused |
| recorded_at | TIMESTAMPTZ NOT NULL | Waktu pencatatan |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | nullable |

Unique: `(activity_session_id, santri_id) WHERE deleted_at IS NULL`

Constants:
```
AttendanceStatus: present, absent, excused
```

## Alur Data

```
Program
    │
    └── Santri (via kesantrian.Contract)
          │
          └── SantriRegistration ──── AcademicPeriod
                                          │
                                          └── ActivityPeriod
                                                 │
Activity ───────────────────────────────────────┘
                                                 │
                                                 ├── ActivityPeriodProgram ── Program
                                                 │
                                                 └── ActivitySchedule
                                                        │
                                                        ├── ActivityScheduleWeekly
                                                        ├── ActivityScheduleMonthly
                                                        └── ActivityScheduleYearly
                                                               │
                                                               ▼
                                                        ActivitySession
                                                               │
                                                               ▼
                                                           Attendance
                                                               │
                                                               ▼
                                                          Santri (kesantrian)
```

## Cross-Module Dependencies

### Kesantrian → Akademik
Akademik membutuhkan akses data santri:
- `GetSantriByID(santriID)` — validasi santri exists & aktif
- `ListSantriByIDs(santriIDs)` — batch lookup untuk absensi
- `ListActiveSantriByProgram(programID)` — list santri aktif per program

Menggunakan `kesantrian.Contract` yang sudah ada (tambah method jika perlu).

### Pattern
- `internal/modules/akademik/application/ports/kesantrian_reader.go` — port
- `internal/modules/akademik/infrastructure/kesantriangateway/gateway.go` — adapter
- `akademik.NewModule(..., kesantrianContract kesantrian.Contract)` — wiring

## Struktur Module

```
internal/modules/akademik/
  module.go
  contract.go                              -- Contract interface (jika diperlukan module lain)
  domain/
    program/
      entity/program.go
      constant/program_constant.go
      repository/interfaces.go
    academic_period/
      entity/academic_period.go
      constant/academic_period_constant.go
      repository/interfaces.go
    santri_registration/
      entity/santri_registration.go
      constant/santri_registration_constant.go
      repository/interfaces.go
    activity/
      entity/activity.go
      constant/activity_constant.go
      repository/interfaces.go
    activity_period/
      entity/activity_period.go
      constant/activity_period_constant.go
      repository/interfaces.go
    activity_period_program/
      entity/activity_period_program.go
      repository/interfaces.go
    activity_schedule/
      entity/activity_schedule.go
      constant/activity_schedule_constant.go
      repository/interfaces.go
    activity_session/
      entity/activity_session.go
      constant/activity_session_constant.go
      repository/interfaces.go
    attendance/
      entity/attendance.go
      constant/attendance_constant.go
      repository/interfaces.go
  application/
    command/
      -- program
      create_program.go
      update_program.go
      -- academic_period
      create_academic_period.go
      update_academic_period.go
      close_academic_period.go
      -- santri_registration
      register_santri.go
      complete_registration.go
      cancel_registration.go
      -- activity
      create_activity.go
      update_activity.go
      -- activity_period
      activate_activity_period.go
      deactivate_activity_period.go
      -- activity_period_program
      assign_program_to_activity_period.go
      remove_program_from_activity_period.go
      -- activity_schedule
      create_schedule.go
      update_schedule.go
      delete_schedule.go
      -- activity_session
      create_session.go
      cancel_session.go
      complete_session.go
      -- attendance
      record_attendance.go
      update_attendance.go
    query/
      list_programs.go
      get_program.go
      list_academic_periods.go
      get_academic_period.go
      list_santri_registrations.go
      get_santri_registration.go
      list_activities.go
      get_activity.go
      list_activity_periods.go
      list_activity_schedules.go
      get_activity_schedule.go
      list_activity_sessions.go
      list_attendance.go
    dto/
      program_dto.go
      academic_period_dto.go
      santri_registration_dto.go
      activity_dto.go
      activity_period_dto.go
      activity_schedule_dto.go
      activity_session_dto.go
      attendance_dto.go
    ports/
      kesantrian_reader.go
      transactor.go
    errors.go
  infrastructure/
    persistence/
      postgres_program_repo.go
      postgres_academic_period_repo.go
      postgres_santri_registration_repo.go
      postgres_activity_repo.go
      postgres_activity_period_repo.go
      postgres_activity_period_program_repo.go
      postgres_activity_schedule_repo.go
      postgres_activity_session_repo.go
      postgres_attendance_repo.go
      postgres_transactor.go
      helpers.go
    kesantriangateway/
      gateway.go
  interfaces/http/
    handler.go
    router.go
```

## Migration

Migration baru: `migrations/YYYYMMDDHHMMSS_create_akademik_tables.up.sql` / `.down.sql`

Tabel yang dibuat:
1. `programs`
2. `academic_periods`
3. `santri_registrations`
4. `activities`
5. `activity_periods`
6. `activity_period_programs`
7. `activity_schedules`
8. `activity_schedule_weeklies`
9. `activity_schedule_monthlies`
10. `activity_schedule_yearlies`
11. `activity_sessions`
12. `attendances`

## Permission Keys Baru

Ditambahkan di `internal/modules/identity/domain/role/constant/permission_constant.go`:

- `PermissionManageAkademik = "manage_akademik"` — CRUD program, period, registration,
  activity, schedule, session, attendance (operasional harian)
- `PermissionManageAkademikSettings = "manage_akademik_settings"` — kelola academic period
  lifecycle (open/close/archive) dan konfigurasi master

## Ringkasan Endpoint HTTP (`/api/v1/web/akademik/...`)

Semua endpoint butuh JWT + `manage_akademik` kecuali disebutkan lain.

### Program
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/programs` | List program |
| GET | `/programs/:id` | Detail program |
| POST | `/programs` | Buat program |
| PUT | `/programs/:id` | Update program |

### Academic Period
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/periods` | List academic periods |
| GET | `/periods/:id` | Detail period |
| POST | `/periods` | Buat period |
| PUT | `/periods/:id` | Update period |
| POST | `/periods/:id/open` | Buka period |
| POST | `/periods/:id/close` | Tutup period |

### Santri Registration
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/registrations` | List registrations (filter by period, santri) |
| GET | `/registrations/:id` | Detail registration |
| POST | `/registrations` | Registrasi santri ke period |
| POST | `/registrations/:id/complete` | Selesaikan registrasi |
| POST | `/registrations/:id/cancel` | Batalkan registrasi |

### Activity
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/activities` | List activities |
| GET | `/activities/:id` | Detail activity |
| POST | `/activities` | Buat activity |
| PUT | `/activities/:id` | Update activity |

### Activity Period
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/activity-periods` | List activity periods (filter by period, activity) |
| POST | `/activity-periods` | Aktivasi activity di period |
| POST | `/activity-periods/:id/activate` | Aktifkan |
| POST | `/activity-periods/:id/deactivate` | Nonaktifkan |

### Activity Period Program
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/activity-periods/:id/programs` | List program scope |
| POST | `/activity-periods/:id/programs` | Assign program |
| DELETE | `/activity-periods/:id/programs/:programId` | Hapus program scope |

### Activity Schedule
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/activity-periods/:id/schedules` | List schedules per period |
| GET | `/schedules/:id` | Detail schedule + recurrence rules |
| POST | `/schedules` | Buat schedule |
| PUT | `/schedules/:id` | Update schedule |
| DELETE | `/schedules/:id` | Hapus schedule |

### Activity Session
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/sessions` | List sessions (filter by schedule, period, date range) |
| GET | `/sessions/:id` | Detail session |
| POST | `/sessions` | Buat session (manual / dari schedule) |
| POST | `/sessions/:id/cancel` | Cancel session |
| POST | `/sessions/:id/complete` | Complete session |

### Attendance
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/sessions/:id/attendance` | List attendance per session |
| POST | `/sessions/:id/attendance` | Record attendance (batch) |
| PUT | `/sessions/:id/attendance/:santriId` | Update attendance status |

## Verifikasi

1. `go build ./...` lolos.
2. Migration `up`/`down` bersih di DB lokal.
3. Smoke test E2E:
   - Buat program TAHFIDZ & KITAB
   - Buat academic period, set open
   - Registrasi santri ke period, complete
   - Buat activity (Shalat Subuh, Kajian)
   - Aktivasi activity di period
   - Assign program scope
   - Buat schedule (daily untuk shalat, weekly untuk kajian)
   - Buat session dari schedule
   - Record attendance
   - Verify query: siapa santri di program apa, period mana aktif, kegiatan
     apa saja, session kapan, santri hadir atau tidak
4. Test constraint: duplicate registration ditolak, duplicate attendance ditolak.
5. Test permission: tanpa `manage_akademik` → 403.

## Candidate Constants Summary

```
ProgramCode:        tahfidz, kitab
ProgramStatus:      active, inactive

AcademicPeriodStatus: draft, open, closed, archived

SantriRegistrationStatus: pending, completed, cancelled

ActivityStatus:     active, inactive
ActivityPeriodStatus: active, inactive

ActivityScheduleType: once, daily, weekly, monthly, yearly
DayOfWeek:          monday, tuesday, wednesday, thursday, friday, saturday, sunday

ActivitySessionStatus: scheduled, open, completed, cancelled
AttendanceStatus:   present, absent, excused
```

Semua constant value lowercase (mengikuti pola existing: `draft`, `active`, `pending`, dst).
Error codes: UPPER_SNAKE_CASE dengan prefix domain (e.g. `PROGRAM_NOT_FOUND`,
`ACTIVITY_PERIOD_DUPLICATE`).
