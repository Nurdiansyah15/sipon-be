-- ============================================================
-- Relasi periode tagihan → periode akuntansi (hierarchy)
-- ============================================================

ALTER TABLE billing_periods ADD COLUMN accounting_period_id UUID;

-- Backfill data lama bila ada: pasangkan ke periode akuntansi yang
-- rentang tanggalnya memuat periode tagihan.
UPDATE billing_periods bp
SET accounting_period_id = (
    SELECT ap.id
    FROM accounting_periods ap
    WHERE ap.start_date <= bp.start_date AND ap.end_date >= bp.end_date
    LIMIT 1
);

ALTER TABLE billing_periods ALTER COLUMN accounting_period_id SET NOT NULL;

CREATE INDEX idx_billing_periods_accounting_period ON billing_periods(accounting_period_id);

-- Validasi rentang tanggal periode tagihan berada di dalam periode akuntansi
-- dilakukan di level aplikasi/domain (bukan di DB).
