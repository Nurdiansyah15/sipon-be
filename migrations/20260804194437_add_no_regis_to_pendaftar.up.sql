ALTER TABLE pendaftar ADD COLUMN no_regis VARCHAR(15);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pendaftar_no_regis ON pendaftar (no_regis) WHERE no_regis IS NOT NULL;
