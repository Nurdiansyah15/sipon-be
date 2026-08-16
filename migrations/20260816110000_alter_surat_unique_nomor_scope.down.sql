ALTER TABLE surat DROP CONSTRAINT IF EXISTS uq_surat_nomor_scope;
ALTER TABLE surat ADD CONSTRAINT surat_nomor_key UNIQUE (nomor);
