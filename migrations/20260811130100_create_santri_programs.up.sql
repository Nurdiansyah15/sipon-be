-- Pemetaan santri ke program (akademik).
-- Santri bisa punya banyak program historis, tapi hanya 1 yang aktif.
-- santri_id plain UUID tanpa FK — cross-module (kesantrian).

CREATE TABLE IF NOT EXISTS santri_programs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_id   UUID NOT NULL,
    program_id  UUID NOT NULL REFERENCES programs(id),
    is_active   BOOLEAN NOT NULL DEFAULT true,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Hanya 1 program aktif per santri.
CREATE UNIQUE INDEX IF NOT EXISTS idx_santri_programs_active_unique
    ON santri_programs(santri_id)
    WHERE is_active = true AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_santri_programs_santri
    ON santri_programs(santri_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_santri_programs_program
    ON santri_programs(program_id)
    WHERE deleted_at IS NULL;
