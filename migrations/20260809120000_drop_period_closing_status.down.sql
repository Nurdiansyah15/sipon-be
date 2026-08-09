ALTER TABLE accounting_periods DROP CONSTRAINT IF EXISTS accounting_periods_status_check;

ALTER TABLE accounting_periods ADD CONSTRAINT accounting_periods_status_check
    CHECK (status IN ('open', 'closing', 'closed', 'locked'));
