DROP INDEX IF EXISTS idx_invoices_unique_periode;
CREATE UNIQUE INDEX idx_invoices_unique_periode
    ON invoices(santri_id, fee_component_id, periode)
    WHERE deleted_at IS NULL AND status NOT IN ('cancelled');

DROP TABLE IF EXISTS invoice_number_counters;
