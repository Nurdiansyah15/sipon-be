CREATE TABLE IF NOT EXISTS dokumen_aset (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    judul VARCHAR(255) NOT NULL,
    deskripsi TEXT,
    kategori VARCHAR(50) NOT NULL,
    key TEXT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_dokumen_aset_kategori ON dokumen_aset(kategori) WHERE deleted_at IS NULL;
CREATE INDEX idx_dokumen_aset_is_public ON dokumen_aset(is_public) WHERE deleted_at IS NULL;
CREATE INDEX idx_dokumen_aset_created_at ON dokumen_aset(created_at DESC) WHERE deleted_at IS NULL;
