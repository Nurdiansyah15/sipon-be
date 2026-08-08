CREATE TABLE IF NOT EXISTS billing_batches (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(200) NOT NULL,
    billing_scheme_id UUID NOT NULL REFERENCES billing_schemes(id),
    billing_period_id UUID NOT NULL REFERENCES billing_periods(id),
    status            VARCHAR(20) NOT NULL DEFAULT 'processing'
                      CHECK (status IN ('processing', 'completed', 'failed')),
    created_by        UUID NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ,
    total_created     INTEGER NOT NULL DEFAULT 0,
    total_skipped     INTEGER NOT NULL DEFAULT 0,
    total_error       INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_billing_batches_status ON billing_batches(status);
CREATE INDEX idx_billing_batches_period ON billing_batches(billing_period_id);

CREATE TABLE IF NOT EXISTS billing_batch_targets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id     UUID NOT NULL REFERENCES billing_batches(id) ON DELETE CASCADE,
    santri_id    UUID NOT NULL,
    status       VARCHAR(30) NOT NULL DEFAULT 'pending'
                 CHECK (status IN (
                    'pending', 'created', 'skipped_no_assignment', 'skipped_wrong_scheme',
                    'skipped_already_invoiced', 'skipped_component_missing', 'error'
                 )),
    invoice_id   UUID REFERENCES invoices(id),
    reason       TEXT,
    processed_at TIMESTAMPTZ,
    UNIQUE(batch_id, santri_id)
);
CREATE INDEX idx_bbt_batch ON billing_batch_targets(batch_id);
CREATE INDEX idx_bbt_status ON billing_batch_targets(batch_id, status);
