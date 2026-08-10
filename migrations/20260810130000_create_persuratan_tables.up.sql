CREATE TABLE IF NOT EXISTS master_tipe_surat (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama        VARCHAR(200) NOT NULL,
    kode        VARCHAR(20) UNIQUE NOT NULL,
    created_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS surat (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor           VARCHAR(100) UNIQUE NOT NULL,
    seq             INT NOT NULL,
    tipe_surat_id   UUID NOT NULL REFERENCES master_tipe_surat(id),
    keterangan      TEXT,
    tanggal         DATE NOT NULL,
    created_by      UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_surat_tipe ON surat(tipe_surat_id);
CREATE INDEX idx_surat_tanggal ON surat(EXTRACT(MONTH FROM tanggal), EXTRACT(YEAR FROM tanggal));

CREATE TABLE IF NOT EXISTS surat_dokumen_aset (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surat_id         UUID NOT NULL REFERENCES surat(id) ON DELETE CASCADE,
    dokumen_aset_id  UUID NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (surat_id, dokumen_aset_id)
);
CREATE INDEX idx_surat_dokumen_surat ON surat_dokumen_aset(surat_id);
CREATE INDEX idx_surat_dokumen_aset ON surat_dokumen_aset(dokumen_aset_id);
