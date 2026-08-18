CREATE TABLE message_jobs (
    id              UUID PRIMARY KEY,
    routing_key     VARCHAR(150) NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    version         INT NOT NULL,
    correlation_id  VARCHAR(64),
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING', 'RUNNING', 'RETRY_WAIT', 'SUCCEEDED', 'FAILED')),
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

CREATE INDEX idx_message_jobs_claim
    ON message_jobs (status, next_attempt_at)
    WHERE status IN ('PENDING', 'RETRY_WAIT');

CREATE INDEX idx_message_jobs_running_lease
    ON message_jobs (status, locked_until)
    WHERE status = 'RUNNING';
