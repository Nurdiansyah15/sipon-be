CREATE TABLE IF NOT EXISTS billing_periods (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    period_type VARCHAR(20) NOT NULL CHECK (period_type IN ('monthly', 'semesterly', 'yearly', 'once', 'weekly')),
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'draft'
                CHECK (status IN ('draft', 'open', 'closed')),
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_billing_periods_status ON billing_periods(status);
CREATE INDEX idx_billing_periods_date_range ON billing_periods(start_date, end_date);

-- No production data yet, so the old free-text periode/tahun_ajaran pair is
-- dropped outright rather than kept alongside the new FK (see
-- docs/plan/perubahan-skema-tagihan.md Fase 1).
DROP INDEX IF EXISTS idx_invoices_unique_periode;
DROP INDEX IF EXISTS idx_invoices_periode;
DROP INDEX IF EXISTS idx_invoices_tahun_ajaran;

ALTER TABLE invoices ADD COLUMN billing_period_id UUID NOT NULL REFERENCES billing_periods(id);
ALTER TABLE invoices DROP COLUMN periode;
ALTER TABLE invoices DROP COLUMN tahun_ajaran;

CREATE INDEX idx_invoices_billing_period ON invoices(billing_period_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_invoices_unique_period
    ON invoices(santri_id, fee_component_id, billing_period_id)
    WHERE deleted_at IS NULL AND status NOT IN ('cancelled');
