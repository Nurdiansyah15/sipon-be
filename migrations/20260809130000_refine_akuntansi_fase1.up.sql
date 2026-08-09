-- Fase 1 penyempurnaan akuntansi.
-- 1) payments.debit_account_id jadi NOT NULL (wajib akun kas/bank tujuan
--    sebelum payment bisa diverifikasi & diposting ke jurnal).
-- 2) Unique partial index anti double-posting jurnal otomatis.

-- Backfill baris lama tanpa debit_account_id ke akun kas default (1101).
UPDATE payments SET debit_account_id = (SELECT id FROM accounts WHERE code = '1101')
WHERE debit_account_id IS NULL;

ALTER TABLE payments ALTER COLUMN debit_account_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_journal_source_unique
    ON journal_entries(source_type, source_id)
    WHERE source_type IS NOT NULL AND source_type != 'manual' AND status != 'cancelled';
