ALTER TABLE activity_schedules
    DROP COLUMN IF EXISTS early_minutes,
    DROP COLUMN IF EXISTS late_minutes;
