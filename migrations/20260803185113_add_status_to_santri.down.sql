ALTER TABLE santri
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS status_changed_by,
    DROP COLUMN IF EXISTS status_changed_at,
    DROP COLUMN IF EXISTS status_notes;
