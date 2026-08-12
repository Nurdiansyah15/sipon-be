-- Rollback status 'draft'.

ALTER TABLE santri_registrations DROP CONSTRAINT IF EXISTS santri_registrations_status_check;
ALTER TABLE santri_registrations ADD CONSTRAINT santri_registrations_status_check
    CHECK (status IN ('pending', 'revision', 'completed', 'cancelled'));
