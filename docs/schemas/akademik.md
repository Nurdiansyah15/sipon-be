# Schema: Akademik

Dokumen ini merangkum skema DB modul akademik.
Migration: `create_akademik_tables`.

---

## `programs` — master program pendidikan

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| code | VARCHAR(20) UNIQUE NOT NULL | e.g. TAHFIDZ, KITAB |
| name | VARCHAR(100) NOT NULL | Nama program |
| status | VARCHAR(20) NOT NULL DEFAULT 'active' | CHECK: active, inactive |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| deleted_at | TIMESTAMPTZ | nullable, soft delete |

Index:
- `idx_programs_code` UNIQUE pada `code` (implicit)
- `idx_programs_active` partial: `WHERE deleted_at IS NULL`

---

## `academic_periods` — periode akademik

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| code | VARCHAR(50) UNIQUE NOT NULL | e.g. 2026/2027-P1 |
| name | VARCHAR(100) NOT NULL | e.g. Periode 1 2026/2027 |
| start_date | DATE NOT NULL | |
| end_date | DATE NOT NULL | |
| status | VARCHAR(20) NOT NULL DEFAULT 'draft' | CHECK: draft, open, closed, archived |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| deleted_at | TIMESTAMPTZ | nullable |

Constraint:
- `CHECK (end_date >= start_date)`

Index:
- `idx_academic_periods_code` UNIQUE pada `code` (implicit)
- `idx_academic_periods_status` pada `status WHERE deleted_at IS NULL`
- `idx_academic_periods_date_range` pada `(start_date, end_date)`

---

## `santri_registrations` — herregistrasi santri per periode

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| santri_id | UUID NOT NULL | Cross-module ke kesantrian (no FK) |
| academic_period_id | UUID NOT NULL FK → academic_periods | |
| status | VARCHAR(20) NOT NULL DEFAULT 'pending' | CHECK: pending, completed, cancelled |
| registered_at | TIMESTAMPTZ | Waktu registrasi diselesaikan |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| deleted_at | TIMESTAMPTZ | nullable |

Constraints:
- `UNIQUE (santri_id, academic_period_id) WHERE deleted_at IS NULL` (partial unique)

Index:
- `idx_santri_registrations_santri` pada `santri_id WHERE deleted_at IS NULL`
- `idx_santri_registrations_period` pada `academic_period_id WHERE deleted_at IS NULL`
- `idx_santri_registrations_status` pada `status WHERE deleted_at IS NULL`

---

## `activities` — master kegiatan

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| code | VARCHAR(50) UNIQUE NOT NULL | e.g. SHALAT_SUBUH, KAJIAN |
| name | VARCHAR(200) NOT NULL | Nama kegiatan |
| status | VARCHAR(20) NOT NULL DEFAULT 'active' | CHECK: active, inactive |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| deleted_at | TIMESTAMPTZ | nullable |

Index:
- `idx_activities_code` UNIQUE pada `code` (implicit)
- `idx_activities_active` partial: `WHERE deleted_at IS NULL`

---

## `activity_periods` — aktivasi activity per academic period

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| activity_id | UUID NOT NULL FK → activities | |
| academic_period_id | UUID NOT NULL FK → academic_periods | |
| status | VARCHAR(20) NOT NULL DEFAULT 'active' | CHECK: active, inactive |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| deleted_at | TIMESTAMPTZ | nullable |

Constraints:
- `UNIQUE (activity_id, academic_period_id) WHERE deleted_at IS NULL`

Index:
- `idx_activity_periods_activity` pada `activity_id WHERE deleted_at IS NULL`
- `idx_activity_periods_period` pada `academic_period_id WHERE deleted_at IS NULL`
- `idx_activity_periods_status` pada `status WHERE deleted_at IS NULL`

---

## `activity_period_programs` — scope program per activity-period

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| activity_period_id | UUID NOT NULL FK → activity_periods ON DELETE CASCADE | |
| program_id | UUID NOT NULL FK → programs | |

Constraints:
- `UNIQUE (activity_period_id, program_id)`

Index:
- `idx_app_activity_period` pada `activity_period_id`
- `idx_app_program` pada `program_id`

**Catatan:** Jika activity_period **tidak punya** record di tabel ini,
activity berlaku untuk **semua program** (default behavior).

---

## `activity_schedules` — pola recurrence kegiatan

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| activity_period_id | UUID NOT NULL FK → activity_periods | |
| type | VARCHAR(20) NOT NULL | CHECK: once, daily, weekly, monthly, yearly |
| start_date | DATE | Rentang mulai berlakunya schedule |
| end_date | DATE | Rentang akhir berlakunya schedule |
| start_time | TIME NOT NULL | Waktu mulai |
| end_time | TIME NOT NULL | Waktu selesai |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| deleted_at | TIMESTAMPTZ | nullable |

Constraints:
- `CHECK (end_time > start_time)`
- `CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)`

Index:
- `idx_schedules_period` pada `activity_period_id WHERE deleted_at IS NULL`
- `idx_schedules_type` pada `type WHERE deleted_at IS NULL`

---

## `activity_schedule_weeklies` — detail hari untuk schedule WEEKLY

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| schedule_id | UUID NOT NULL FK → activity_schedules ON DELETE CASCADE | |
| day_of_week | VARCHAR(10) NOT NULL | CHECK: monday-sunday |

Constraints:
- `UNIQUE (schedule_id, day_of_week)`

Index:
- `idx_schedule_weekly_schedule` pada `schedule_id`

---

## `activity_schedule_monthlies` — detail tanggal untuk schedule MONTHLY

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| schedule_id | UUID NOT NULL FK → activity_schedules ON DELETE CASCADE | |
| day_of_month | INT NOT NULL | 1-31 |

Constraints:
- `CHECK (day_of_month BETWEEN 1 AND 31)`
- `UNIQUE (schedule_id, day_of_month)`

Index:
- `idx_schedule_monthly_schedule` pada `schedule_id`

---

## `activity_schedule_yearlies` — detail bulan+tanggal untuk schedule YEARLY

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| schedule_id | UUID NOT NULL FK → activity_schedules ON DELETE CASCADE | |
| month | INT NOT NULL | 1-12 |
| day | INT NOT NULL | 1-31 |

Constraints:
- `CHECK (month BETWEEN 1 AND 12)`
- `CHECK (day BETWEEN 1 AND 31)`
- `UNIQUE (schedule_id, month, day)`

Index:
- `idx_schedule_yearly_schedule` pada `schedule_id`

---

## `activity_sessions` — occurrence konkret kegiatan

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| activity_schedule_id | UUID NOT NULL FK → activity_schedules | |
| starts_at | TIMESTAMPTZ NOT NULL | Waktu mulai session |
| ends_at | TIMESTAMPTZ NOT NULL | Waktu selesai session |
| status | VARCHAR(20) NOT NULL DEFAULT 'scheduled' | CHECK: scheduled, open, completed, cancelled |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| deleted_at | TIMESTAMPTZ | nullable |

Constraints:
- `CHECK (ends_at > starts_at)`

Index:
- `idx_sessions_schedule` pada `activity_schedule_id WHERE deleted_at IS NULL`
- `idx_sessions_time` pada `(starts_at, ends_at) WHERE deleted_at IS NULL`
- `idx_sessions_status` pada `status WHERE deleted_at IS NULL`

---

## `attendances` — absensi santri per session

| Kolom | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| activity_session_id | UUID NOT NULL FK → activity_sessions | |
| santri_id | UUID NOT NULL | Cross-module ke kesantrian (no FK) |
| status | VARCHAR(20) NOT NULL | CHECK: present, absent, excused |
| recorded_at | TIMESTAMPTZ NOT NULL | Waktu pencatatan absensi |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| deleted_at | TIMESTAMPTZ | nullable |

Constraints:
- `UNIQUE (activity_session_id, santri_id) WHERE deleted_at IS NULL`

Index:
- `idx_attendances_session` pada `activity_session_id WHERE deleted_at IS NULL`
- `idx_attendances_santri` pada `santri_id WHERE deleted_at IS NULL`
- `idx_attendances_recorded` pada `recorded_at`

---

## DDL

```sql
-- programs
CREATE TABLE IF NOT EXISTS programs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(20) UNIQUE NOT NULL,
    name        VARCHAR(100) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'inactive')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_programs_active ON programs(code) WHERE deleted_at IS NULL;

-- academic_periods
CREATE TABLE IF NOT EXISTS academic_periods (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50) UNIQUE NOT NULL,
    name        VARCHAR(100) NOT NULL,
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'draft'
                CHECK (status IN ('draft', 'open', 'closed', 'archived')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    CHECK (end_date >= start_date)
);
CREATE INDEX idx_academic_periods_status ON academic_periods(status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_academic_periods_date_range ON academic_periods(start_date, end_date);

-- santri_registrations
CREATE TABLE IF NOT EXISTS santri_registrations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_id           UUID NOT NULL,
    academic_period_id  UUID NOT NULL REFERENCES academic_periods(id),
    status              VARCHAR(20) NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'completed', 'cancelled')),
    registered_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_santri_registrations_unique
    ON santri_registrations(santri_id, academic_period_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_santri_registrations_santri ON santri_registrations(santri_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_santri_registrations_period ON santri_registrations(academic_period_id)
    WHERE deleted_at IS NULL;

-- activities
CREATE TABLE IF NOT EXISTS activities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50) UNIQUE NOT NULL,
    name        VARCHAR(200) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'inactive')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_activities_active ON activities(code) WHERE deleted_at IS NULL;

-- activity_periods
CREATE TABLE IF NOT EXISTS activity_periods (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id         UUID NOT NULL REFERENCES activities(id),
    academic_period_id  UUID NOT NULL REFERENCES academic_periods(id),
    status              VARCHAR(20) NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'inactive')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_activity_periods_unique
    ON activity_periods(activity_id, academic_period_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_periods_activity ON activity_periods(activity_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_activity_periods_period ON activity_periods(academic_period_id)
    WHERE deleted_at IS NULL;

-- activity_period_programs
CREATE TABLE IF NOT EXISTS activity_period_programs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_period_id  UUID NOT NULL REFERENCES activity_periods(id) ON DELETE CASCADE,
    program_id          UUID NOT NULL REFERENCES programs(id),
    UNIQUE (activity_period_id, program_id)
);
CREATE INDEX idx_app_activity_period ON activity_period_programs(activity_period_id);
CREATE INDEX idx_app_program ON activity_period_programs(program_id);

-- activity_schedules
CREATE TABLE IF NOT EXISTS activity_schedules (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_period_id  UUID NOT NULL REFERENCES activity_periods(id),
    type                VARCHAR(20) NOT NULL
                        CHECK (type IN ('once', 'daily', 'weekly', 'monthly', 'yearly')),
    start_date          DATE,
    end_date            DATE,
    start_time          TIME NOT NULL,
    end_time            TIME NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    CHECK (end_time > start_time),
    CHECK (end_date IS NULL OR start_date IS NULL OR end_date >= start_date)
);
CREATE INDEX idx_schedules_period ON activity_schedules(activity_period_id)
    WHERE deleted_at IS NULL;

-- activity_schedule_weeklies
CREATE TABLE IF NOT EXISTS activity_schedule_weeklies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES activity_schedules(id) ON DELETE CASCADE,
    day_of_week     VARCHAR(10) NOT NULL
                    CHECK (day_of_week IN (
                        'monday', 'tuesday', 'wednesday', 'thursday',
                        'friday', 'saturday', 'sunday'
                    )),
    UNIQUE (schedule_id, day_of_week)
);
CREATE INDEX idx_schedule_weekly_schedule ON activity_schedule_weeklies(schedule_id);

-- activity_schedule_monthlies
CREATE TABLE IF NOT EXISTS activity_schedule_monthlies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES activity_schedules(id) ON DELETE CASCADE,
    day_of_month    INT NOT NULL CHECK (day_of_month BETWEEN 1 AND 31),
    UNIQUE (schedule_id, day_of_month)
);
CREATE INDEX idx_schedule_monthly_schedule ON activity_schedule_monthlies(schedule_id);

-- activity_schedule_yearlies
CREATE TABLE IF NOT EXISTS activity_schedule_yearlies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES activity_schedules(id) ON DELETE CASCADE,
    month           INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    day             INT NOT NULL CHECK (day BETWEEN 1 AND 31),
    UNIQUE (schedule_id, month, day)
);
CREATE INDEX idx_schedule_yearly_schedule ON activity_schedule_yearlies(schedule_id);

-- activity_sessions
CREATE TABLE IF NOT EXISTS activity_sessions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_schedule_id    UUID NOT NULL REFERENCES activity_schedules(id),
    starts_at               TIMESTAMPTZ NOT NULL,
    ends_at                 TIMESTAMPTZ NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'scheduled'
                            CHECK (status IN ('scheduled', 'open', 'completed', 'cancelled')),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    CHECK (ends_at > starts_at)
);
CREATE INDEX idx_sessions_schedule ON activity_sessions(activity_schedule_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_time ON activity_sessions(starts_at, ends_at)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_sessions_status ON activity_sessions(status)
    WHERE deleted_at IS NULL;

-- attendances
CREATE TABLE IF NOT EXISTS attendances (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_session_id UUID NOT NULL REFERENCES activity_sessions(id),
    santri_id           UUID NOT NULL,
    status              VARCHAR(20) NOT NULL
                        CHECK (status IN ('present', 'absent', 'excused')),
    recorded_at         TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_attendances_unique
    ON attendances(activity_session_id, santri_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_attendances_session ON attendances(activity_session_id)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_attendances_santri ON attendances(santri_id)
    WHERE deleted_at IS NULL;
```
