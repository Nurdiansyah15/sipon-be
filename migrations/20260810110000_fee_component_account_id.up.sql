-- Ganti pemetaan komponen biaya dari enum `type` ke referensi akun COA langsung.
-- Lihat docs/plan/fee-component-account-id.md.
-- Prasyarat: accounts.sub_type sudah ada (migrasi 20260810100000_add_account_sub_type).
-- Catatan: belum ada data production nyata di fee_components, jadi kolom baru
-- langsung NOT NULL tanpa langkah backfill, dan kolom `type` langsung di-drop.

ALTER TABLE fee_components
    ADD COLUMN revenue_account_id UUID NOT NULL REFERENCES accounts(id),
    ADD COLUMN receivable_account_id UUID NOT NULL REFERENCES accounts(id);

DROP INDEX IF EXISTS idx_fee_components_type;
ALTER TABLE fee_components DROP COLUMN type;

CREATE INDEX idx_fee_components_revenue_account ON fee_components(revenue_account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_fee_components_receivable_account ON fee_components(receivable_account_id) WHERE deleted_at IS NULL;
