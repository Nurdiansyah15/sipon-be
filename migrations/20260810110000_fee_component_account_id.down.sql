DROP INDEX IF EXISTS idx_fee_components_revenue_account;
DROP INDEX IF EXISTS idx_fee_components_receivable_account;

ALTER TABLE fee_components ADD COLUMN type VARCHAR(30) CHECK (type IN ('ukt', 'spp', 'daftar_ulang', 'insidental'));
CREATE INDEX idx_fee_components_type ON fee_components(type) WHERE deleted_at IS NULL;

ALTER TABLE fee_components
    DROP COLUMN revenue_account_id,
    DROP COLUMN receivable_account_id;
