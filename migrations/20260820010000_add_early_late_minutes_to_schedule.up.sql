ALTER TABLE activity_schedules
    ADD COLUMN early_minutes INT NOT NULL DEFAULT 0,
    ADD COLUMN late_minutes  INT NOT NULL DEFAULT 0;
