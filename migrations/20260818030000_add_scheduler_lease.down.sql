DROP INDEX IF EXISTS idx_scheduled_jobs_processing_lease;

ALTER TABLE scheduled_jobs
    DROP COLUMN IF EXISTS lease_until;
