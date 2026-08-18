ALTER TABLE scheduled_jobs
    ADD COLUMN lease_until TIMESTAMPTZ;

CREATE INDEX idx_scheduled_jobs_processing_lease
    ON scheduled_jobs (status, lease_until)
    WHERE status = 'PROCESSING';
