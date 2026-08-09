-- Tambah klasifikasi halus sub_type pada Chart of Accounts (COA).
-- Lihat docs/plan/coa-sub-type.md untuk taksonomi lengkap.

ALTER TABLE accounts ADD COLUMN sub_type VARCHAR(40) CHECK (sub_type IN (
    'cash_bank','receivable','prepaid_expense','inventory','fixed_asset',
    'accumulated_depreciation','intangible_asset','investment','other_asset',
    'payable','customer_advance','unearned_revenue','tax_payable',
    'accrued_liability','long_term_liability','other_liability',
    'capital','retained_earnings','current_year_earnings','withdrawal',
    'operating_revenue','non_operating_revenue',
    'cost_of_goods_sold','operating_expense','depreciation_expense',
    'non_operating_expense','tax_expense'
));
CREATE INDEX idx_accounts_sub_type ON accounts(sub_type) WHERE deleted_at IS NULL;

-- Benahi gap: payments.debit_account_id selama ini UUID polos tanpa FK.
ALTER TABLE payments ADD CONSTRAINT fk_payments_debit_account
    FOREIGN KEY (debit_account_id) REFERENCES accounts(id);

-- Backfill akun existing yang postable.
UPDATE accounts SET sub_type = 'cash_bank' WHERE code IN ('1101','1102');
UPDATE accounts SET sub_type = 'receivable' WHERE code = '1103';
UPDATE accounts SET sub_type = 'fixed_asset' WHERE code IN ('1201','1202','1203','1204');
UPDATE accounts SET sub_type = 'payable' WHERE code = '2101';
UPDATE accounts SET sub_type = 'customer_advance' WHERE code = '2102';
UPDATE accounts SET sub_type = 'unearned_revenue' WHERE code = '2103';
UPDATE accounts SET sub_type = 'capital' WHERE code = '3100';
UPDATE accounts SET sub_type = 'retained_earnings' WHERE code = '3200';
UPDATE accounts SET sub_type = 'current_year_earnings' WHERE code = '3201';
UPDATE accounts SET sub_type = 'operating_revenue' WHERE code IN ('4100','4200','4300','4400');
UPDATE accounts SET sub_type = 'non_operating_revenue' WHERE code IN ('4500','4600');
UPDATE accounts SET sub_type = 'operating_expense' WHERE code IN ('5100','5200','5300');
UPDATE accounts SET sub_type = 'depreciation_expense' WHERE code = '5400';
UPDATE accounts SET sub_type = 'non_operating_expense' WHERE code = '5500';
