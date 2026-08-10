-- Keuangan settings: single-row table menyimpan konfigurasi default keuangan
-- dalam satu kolom JSONB. Key didefinisikan hardcoded di Go constant
-- (domain/setting/constant). Menambah setting baru cukup menambah key di
-- constant + mapper — tidak perlu migration schema baru.
-- Lihat docs/plan/keuangan-settings.md.

CREATE TABLE keuangan_settings (
    id UUID PRIMARY KEY,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Single-row enforcement: selalu pakai ID tetap yang sama.
INSERT INTO keuangan_settings (id, settings)
VALUES ('00000000-0000-0000-0000-000000000001', '{}'::jsonb);
