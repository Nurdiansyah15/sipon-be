CREATE TABLE IF NOT EXISTS scopes (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_type  VARCHAR(50)  NOT NULL,
    code        VARCHAR(100) NOT NULL,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (scope_type, code)
);

CREATE INDEX idx_scopes_type        ON scopes (scope_type);
CREATE INDEX idx_scopes_type_active ON scopes (scope_type) WHERE is_active = TRUE;
