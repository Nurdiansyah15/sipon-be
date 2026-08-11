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

CREATE TABLE IF NOT EXISTS activity_period_programs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_period_id  UUID NOT NULL REFERENCES activity_periods(id) ON DELETE CASCADE,
    program_id          UUID NOT NULL REFERENCES programs(id),
    UNIQUE (activity_period_id, program_id)
);
CREATE INDEX idx_app_activity_period ON activity_period_programs(activity_period_id);
CREATE INDEX idx_app_program ON activity_period_programs(program_id);

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

CREATE TABLE IF NOT EXISTS activity_schedule_monthlies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES activity_schedules(id) ON DELETE CASCADE,
    day_of_month    INT NOT NULL CHECK (day_of_month BETWEEN 1 AND 31),
    UNIQUE (schedule_id, day_of_month)
);
CREATE INDEX idx_schedule_monthly_schedule ON activity_schedule_monthlies(schedule_id);

CREATE TABLE IF NOT EXISTS activity_schedule_yearlies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES activity_schedules(id) ON DELETE CASCADE,
    month           INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    day             INT NOT NULL CHECK (day BETWEEN 1 AND 31),
    UNIQUE (schedule_id, month, day)
);
CREATE INDEX idx_schedule_yearly_schedule ON activity_schedule_yearlies(schedule_id);

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
