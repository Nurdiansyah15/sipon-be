DROP INDEX IF EXISTS idx_invoices_unique_period;
DROP INDEX IF EXISTS idx_invoices_billing_period;

ALTER TABLE invoices ADD COLUMN periode VARCHAR(20);
ALTER TABLE invoices ADD COLUMN tahun_ajaran VARCHAR(10);
ALTER TABLE invoices DROP COLUMN billing_period_id;

ALTER TABLE invoices ALTER COLUMN periode SET NOT NULL;
ALTER TABLE invoices ALTER COLUMN tahun_ajaran SET NOT NULL;

CREATE INDEX idx_invoices_tahun_ajaran ON invoices(tahun_ajaran) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_periode ON invoices(periode) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_invoices_unique_periode
    ON invoices(santri_id, fee_component_id, periode, tahun_ajaran)
    WHERE deleted_at IS NULL AND status NOT IN ('cancelled');

DROP TABLE IF EXISTS billing_periods;
