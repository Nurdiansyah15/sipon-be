CREATE TABLE scheduled_jobs (
    id            UUID         PRIMARY KEY,
    type          VARCHAR(100) NOT NULL,
    payload       JSONB        NOT NULL DEFAULT '{}',
    schedule_type VARCHAR(20)  NOT NULL CHECK (schedule_type IN ('ONE_OFF', 'RECURRING')),
    cron_expr     VARCHAR(100),
    run_at        TIMESTAMPTZ,
    next_run_at   TIMESTAMPTZ  NOT NULL,
    last_run_at   TIMESTAMPTZ,
    status        VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'PROCESSING', 'PAUSED', 'COMPLETED', 'FAILED')),
    retry_count   INT          NOT NULL DEFAULT 0,
    max_retry     INT          NOT NULL DEFAULT 3,
    last_error    TEXT,
    reference_id  VARCHAR(255),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scheduled_jobs_due      ON scheduled_jobs (next_run_at, status) WHERE status = 'ACTIVE';
CREATE INDEX idx_scheduled_jobs_type_ref ON scheduled_jobs (type, reference_id)  WHERE reference_id IS NOT NULL;
