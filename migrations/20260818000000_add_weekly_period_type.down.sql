ALTER TABLE fee_components DROP CONSTRAINT IF EXISTS fee_components_period_type_check;
ALTER TABLE fee_components ADD CONSTRAINT fee_components_period_type_check
    CHECK (period_type IN ('monthly', 'semesterly', 'yearly', 'once'));

ALTER TABLE billing_periods DROP CONSTRAINT IF EXISTS billing_periods_period_type_check;
ALTER TABLE billing_periods ADD CONSTRAINT billing_periods_period_type_check
    CHECK (period_type IN ('monthly', 'semesterly', 'yearly', 'once'));