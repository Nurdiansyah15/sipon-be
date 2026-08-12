-- Rollback perluasan herregistrasi.

DROP TABLE IF EXISTS herregistrasi_documents;
DROP TABLE IF EXISTS herregistrasi_document_requirements;

ALTER TABLE santri_registrations DROP CONSTRAINT IF EXISTS santri_registrations_status_check;
ALTER TABLE santri_registrations ADD CONSTRAINT santri_registrations_status_check
    CHECK (status IN ('pending', 'completed', 'cancelled'));

ALTER TABLE santri_registrations DROP COLUMN IF EXISTS revision_notes;
