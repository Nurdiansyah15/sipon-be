CREATE TABLE IF NOT EXISTS psb_settings (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(200) NOT NULL,
    start_period  DATE NOT NULL,
    end_period    DATE NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'closed')),
    quota         JSONB NOT NULL DEFAULT '{}',
    reg_fee       NUMERIC(12,2) NOT NULL DEFAULT 0,
    bank_accounts JSONB NOT NULL DEFAULT '[]',
    data_purged_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS pendaftar (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    psb_setting_id  UUID NOT NULL REFERENCES psb_settings (id),
    gender          VARCHAR(2) NOT NULL CHECK (gender IN ('1', '2')),
    program         VARCHAR(50),

    nickname         VARCHAR(100),
    hobby            VARCHAR(200),
    purpose          VARCHAR(200),
    motivation_entry VARCHAR(500),
    pob              VARCHAR(200),
    dob              DATE,
    blood            VARCHAR(5),

    address      TEXT,
    sub_district VARCHAR(200),
    district     VARCHAR(200),
    province     VARCHAR(200),
    postal_code  VARCHAR(10),

    previous_pondok_name    VARCHAR(200),
    previous_pondok_address VARCHAR(200),
    previous_pondok_div     VARCHAR(200),
    previous_pondok_time    VARCHAR(100),

    nik    VARCHAR(20),
    no_kk  VARCHAR(20),
    nisn   VARCHAR(10),
    no_kip VARCHAR(20),
    no_kks VARCHAR(20),
    no_pkh VARCHAR(20),

    workplace  VARCHAR(200),
    department VARCHAR(200),

    home_status VARCHAR(100),

    father          VARCHAR(200),
    father_pn       VARCHAR(20),
    father_nik      VARCHAR(20),
    father_job      VARCHAR(200),
    father_graduate VARCHAR(200),
    father_income   VARCHAR(50),

    mother          VARCHAR(200),
    mother_pn       VARCHAR(20),
    mother_nik      VARCHAR(20),
    mother_job      VARCHAR(200),
    mother_graduate VARCHAR(200),
    mother_income   VARCHAR(50),

    guardian_relationship VARCHAR(200),
    guardian              VARCHAR(200),
    guardian_pn           VARCHAR(20),
    guardian_nik          VARCHAR(20),
    guardian_job          VARCHAR(200),
    guardian_graduate     VARCHAR(200),
    guardian_income       VARCHAR(50),

    status          VARCHAR(30) NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'diajukan', 'perlu_revisi', 'ditolak', 'diterima',
                                       'mengundurkan_diri', 'daftar_ulang', 'perlu_revisi_daftar_ulang', 'selesai')),
    accepted_by     UUID,
    accepted_at     TIMESTAMPTZ,
    santri_id       UUID,
    nis             VARCHAR(10),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pendaftar_user_setting
    ON pendaftar (user_id, psb_setting_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS pendaftar_dokumen (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pendaftar_id     UUID NOT NULL REFERENCES pendaftar (id) ON DELETE CASCADE,
    stage            VARCHAR(20) NOT NULL CHECK (stage IN ('pendaftaran', 'daftar_ulang')),
    kind             VARCHAR(30) NOT NULL CHECK (kind IN ('surat_pernyataan', 'ktp', 'kk', 'mutasi', 'pembayaran')),
    key              TEXT NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),
    original_filename VARCHAR(500),
    mime_type         VARCHAR(200),
    size              BIGINT,
    notes             TEXT,
    verified_by UUID,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pendaftar_dokumen_stage_kind
    ON pendaftar_dokumen (pendaftar_id, stage, kind)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS pendaftar_reviews (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pendaftar_id  UUID NOT NULL REFERENCES pendaftar (id) ON DELETE CASCADE,
    stage         VARCHAR(20) NOT NULL CHECK (stage IN ('pendaftaran', 'daftar_ulang')),
    action        VARCHAR(20) NOT NULL CHECK (action IN ('perlu_revisi', 'ditolak', 'diterima')),
    notes         TEXT,
    reviewed_by   UUID NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
