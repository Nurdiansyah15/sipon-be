ALTER TABLE santri
    ADD COLUMN status            VARCHAR(20) NOT NULL DEFAULT 'SANTRI' CHECK (status IN ('SANTRI', 'ALUMNI', 'DROP_OUT')),
    ADD COLUMN status_changed_by UUID,
    ADD COLUMN status_changed_at TIMESTAMPTZ,
    ADD COLUMN status_notes      TEXT;
