ALTER TABLE surat DROP CONSTRAINT IF EXISTS surat_nomor_key;
ALTER TABLE surat ADD CONSTRAINT uq_surat_nomor_scope UNIQUE (nomor, scope_id);
