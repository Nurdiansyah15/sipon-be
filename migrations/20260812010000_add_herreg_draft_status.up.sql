-- Tambah status 'draft' untuk herregistrasi (santri mulai mengisi dokumen).

ALTER TABLE santri_registrations DROP CONSTRAINT IF EXISTS santri_registrations_status_check;
ALTER TABLE santri_registrations ADD CONSTRAINT santri_registrations_status_check
    CHECK (status IN ('draft', 'pending', 'revision', 'completed', 'cancelled'));
