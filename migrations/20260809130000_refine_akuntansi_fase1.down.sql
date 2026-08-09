DROP INDEX IF EXISTS idx_journal_source_unique;

ALTER TABLE payments ALTER COLUMN debit_account_id DROP NOT NULL;
