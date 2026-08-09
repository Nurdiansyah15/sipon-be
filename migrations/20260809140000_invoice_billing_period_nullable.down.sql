-- Kembalikan NOT NULL. TIDAK AMAN kalau sudah ada baris dengan
-- billing_period_id IS NULL (tidak ada default yang masuk akal untuk
-- di-backfill). Ini harus menjadi keputusan manual bila dijalankan.
ALTER TABLE invoices ALTER COLUMN billing_period_id SET NOT NULL;
