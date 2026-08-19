-- 20260819110000_create_device_registrations.up.sql

CREATE TABLE device_registrations (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_token TEXT         NOT NULL,
    push_provider  VARCHAR(20)  NOT NULL,
    platform       VARCHAR(20)  NOT NULL,
    device_id      VARCHAR(255),
    device_name    VARCHAR(255),
    device_model   VARCHAR(255),
    os_version     VARCHAR(50),
    app_version    VARCHAR(50),
    timezone       VARCHAR(50),
    active         BOOLEAN      NOT NULL DEFAULT TRUE,
    last_seen_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_device_reg_provider_token UNIQUE (provider_token)
);

CREATE INDEX idx_device_reg_user_id     ON device_registrations(user_id);
CREATE INDEX idx_device_reg_active      ON device_registrations(active);
CREATE INDEX idx_device_reg_user_active ON device_registrations(user_id, active);
