CREATE TABLE event_outbox (
    id              UUID PRIMARY KEY,
    routing_key     VARCHAR(150) NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    version         INT NOT NULL DEFAULT 1,
    correlation_id  VARCHAR(64),
    causation_id    UUID,
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING', 'PUBLISHING', 'PUBLISHED', 'FAILED')),
    attempt_count   INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at       TIMESTAMPTZ,
    published_at    TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_outbox_claim
    ON event_outbox (next_attempt_at, created_at)
    WHERE status IN ('PENDING', 'FAILED');
