-- Perluasan herregistrasi: status revisi + dokumen.

-- 1. Tambah kolom revision_notes di santri_registrations
ALTER TABLE santri_registrations ADD COLUMN revision_notes TEXT;

-- 2. Update CHECK constraint untuk status baru
ALTER TABLE santri_registrations DROP CONSTRAINT IF EXISTS santri_registrations_status_check;
ALTER TABLE santri_registrations ADD CONSTRAINT santri_registrations_status_check
    CHECK (status IN ('pending', 'revision', 'completed', 'cancelled'));

-- 3. Tabel blueprint dokumen per periode
CREATE TABLE IF NOT EXISTS herregistrasi_document_requirements (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    academic_period_id UUID NOT NULL REFERENCES academic_periods(id),
    kind               VARCHAR(50) NOT NULL,
    label              VARCHAR(200) NOT NULL,
    is_required        BOOLEAN NOT NULL DEFAULT true,
    description        TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ,
    UNIQUE (academic_period_id, kind)
);
CREATE INDEX idx_herreg_doc_req_period ON herregistrasi_document_requirements(academic_period_id) WHERE deleted_at IS NULL;

-- 4. Tabel dokumen herregistrasi
CREATE TABLE IF NOT EXISTS herregistrasi_documents (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_registration_id UUID NOT NULL REFERENCES santri_registrations(id),
    kind                   VARCHAR(50) NOT NULL,
    key                    TEXT NOT NULL,
    original_filename      VARCHAR(500),
    mime_type              VARCHAR(200),
    size                   BIGINT DEFAULT 0,
    status                 VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),
    notes                  TEXT,
    verified_by            UUID,
    verified_at            TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at             TIMESTAMPTZ,
    UNIQUE (santri_registration_id, kind)
);
CREATE INDEX idx_herreg_doc_registration ON herregistrasi_documents(santri_registration_id) WHERE deleted_at IS NULL;
