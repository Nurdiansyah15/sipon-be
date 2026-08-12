-- Request transfer program oleh santri (akademik).
-- Santri mengajukan pindah program; admin approve/reject.
-- Hanya 1 request pending per santri.
-- santri_id plain UUID tanpa FK — cross-module (kesantrian).

CREATE TABLE IF NOT EXISTS program_transfer_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_id       UUID NOT NULL,
    from_program_id UUID NOT NULL REFERENCES programs(id),
    to_program_id   UUID NOT NULL REFERENCES programs(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    notes           TEXT,
    admin_notes     TEXT,
    reviewed_by     UUID,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Hanya 1 request pending per santri.
CREATE UNIQUE INDEX IF NOT EXISTS idx_program_transfer_requests_pending_unique
    ON program_transfer_requests(santri_id)
    WHERE status = 'pending' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_program_transfer_requests_santri
    ON program_transfer_requests(santri_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_program_transfer_requests_status
    ON program_transfer_requests(status)
    WHERE deleted_at IS NULL;
