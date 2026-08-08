CREATE TABLE IF NOT EXISTS invoice_number_counters (
    year INTEGER PRIMARY KEY,
    seq  INTEGER NOT NULL DEFAULT 0
);

-- Duplicate-invoice check previously ignored tahun_ajaran, so the same
-- periode string reused across school years was wrongly treated as a dupe.
DROP INDEX IF EXISTS idx_invoices_unique_periode;
CREATE UNIQUE INDEX idx_invoices_unique_periode
    ON invoices(santri_id, fee_component_id, periode, tahun_ajaran)
    WHERE deleted_at IS NULL AND status NOT IN ('cancelled');
