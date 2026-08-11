-- Akademik settings: single-row table untuk konfigurasi default akademik.
-- Pattern sama dengan keuangan_settings. Key didefinisikan hardcoded di Go
-- constant (domain/setting/constant). Menambah setting baru cukup menambah key
-- di constant + mapper — tidak perlu migration schema baru.
-- Lihat docs/plan/santri-program-mapping.md.

CREATE TABLE IF NOT EXISTS akademik_settings (
    id UUID PRIMARY KEY,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Single-row enforcement: selalu pakai ID tetap yang sama.
INSERT INTO akademik_settings (id, settings)
VALUES ('00000000-0000-0000-0000-000000000002', '{}'::jsonb);
