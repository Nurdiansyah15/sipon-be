ALTER TABLE payments DROP CONSTRAINT fk_payments_debit_account;

DROP INDEX IF EXISTS idx_accounts_sub_type;

ALTER TABLE accounts DROP COLUMN sub_type;
