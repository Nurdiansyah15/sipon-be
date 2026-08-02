-- santri: 1:1 with identity's users table. user_id is a plain UUID with NO
-- foreign key — kesantrian is a separate module from identity, and modules
-- must never take a DB-level dependency on another module's schema. User
-- existence/consistency is enforced at the application layer via
-- identity.Contract, not via a FK constraint.
CREATE TABLE IF NOT EXISTS santri (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    nis     VARCHAR(10) UNIQUE,

    nickname         VARCHAR(100),
    program          VARCHAR(100),
    "option"         VARCHAR(2) CHECK ("option" IN ('1', '2')),
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

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_santri_nik ON santri (nik);
CREATE INDEX IF NOT EXISTS idx_santri_nisn ON santri (nisn);

-- santri_dokumen: intra-module FK to santri is fine (both owned by
-- kesantrian). verified_by references identity's users(id) conceptually,
-- but again NO FK — cross-module.
CREATE TABLE IF NOT EXISTS santri_dokumen (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_id UUID NOT NULL REFERENCES santri (id) ON DELETE CASCADE,

    kind   VARCHAR(30) NOT NULL CHECK (kind IN ('surat_pernyataan', 'ktp', 'kk', 'mutasi', 'pembayaran')),
    key    TEXT        NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),

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
CREATE INDEX IF NOT EXISTS idx_santri_dokumen_santri_id ON santri_dokumen (santri_id);
CREATE INDEX IF NOT EXISTS idx_santri_dokumen_kind ON santri_dokumen (kind);
CREATE INDEX IF NOT EXISTS idx_santri_dokumen_status ON santri_dokumen (status);

-- santri_requests: user_id/reviewed_by reference identity's users(id)
-- conceptually, but again NO FK — cross-module.
CREATE TABLE IF NOT EXISTS santri_requests (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    nis     VARCHAR(10),

    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    notes  TEXT,

    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_santri_requests_user_id ON santri_requests (user_id);
CREATE INDEX IF NOT EXISTS idx_santri_requests_status ON santri_requests (status);
-- Enforce one pending request per user at the DB level (app-level guard
-- alone is not race-safe).
CREATE UNIQUE INDEX IF NOT EXISTS idx_santri_requests_user_pending
    ON santri_requests (user_id)
    WHERE status = 'pending' AND deleted_at IS NULL;
