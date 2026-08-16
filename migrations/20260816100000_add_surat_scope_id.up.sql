ALTER TABLE surat ADD COLUMN scope_id UUID REFERENCES scopes(id);
CREATE INDEX idx_surat_scope_id ON surat (scope_id);
