-- 20260819100000_create_notification_tables.up.sql
-- Tables: notifications, delivery_attempts, notification_preferences

-- ── notifications ─────────────────────────────────────────────────────────────
-- Blueprint/intent record — delivery per-user per-channel is in delivery_attempts.
CREATE TABLE notifications (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    type           VARCHAR(50)  NOT NULL,
    title          VARCHAR(255) NOT NULL,
    body           TEXT         NOT NULL,
    payload        JSONB        NOT NULL DEFAULT '{}',
    reference_id   VARCHAR(255),
    reference_type VARCHAR(100),
    audience_type  VARCHAR(20)  NOT NULL DEFAULT 'unicast',
    audience_data  JSONB        NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notif_audience_type ON notifications(audience_type);

-- ── delivery_attempts ─────────────────────────────────────────────────────────
-- One row = one user × one channel for a given notification.
CREATE TABLE delivery_attempts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID        NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id)         ON DELETE CASCADE,
    channel         VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL CHECK (status IN ('pending','success','failed','retrying')),
    provider_code   VARCHAR(100),
    retry_count     INT         NOT NULL DEFAULT 0,
    next_retry_at   TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,
    attempted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_notif_id          ON delivery_attempts(notification_id);
CREATE INDEX idx_delivery_status            ON delivery_attempts(status);
CREATE INDEX idx_delivery_user_id           ON delivery_attempts(user_id);
CREATE INDEX idx_delivery_user_inapp        ON delivery_attempts(user_id, attempted_at DESC) WHERE channel = 'in_app';
CREATE INDEX idx_delivery_user_inapp_unread ON delivery_attempts(user_id) WHERE channel = 'in_app' AND read_at IS NULL;
CREATE INDEX idx_delivery_retry_at          ON delivery_attempts(next_retry_at) WHERE status = 'retrying';

-- ── notification_preferences ──────────────────────────────────────────────────
CREATE TABLE notification_preferences (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    all_enabled        BOOLEAN     NOT NULL DEFAULT TRUE,
    do_not_disturb     BOOLEAN     NOT NULL DEFAULT FALSE,
    dnd_start_time     VARCHAR(5),
    dnd_end_time       VARCHAR(5),
    module_preferences JSONB       NOT NULL DEFAULT '{}',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
