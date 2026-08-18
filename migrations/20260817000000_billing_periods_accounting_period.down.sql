DROP INDEX IF EXISTS idx_billing_periods_accounting_period;
ALTER TABLE billing_periods DROP COLUMN accounting_period_id;
