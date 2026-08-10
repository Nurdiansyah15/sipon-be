-- Pembayaran santri (self-service) dibuat dengan status pending TANPA
-- debit_account_id. debit_account_id wajib diisi admin saat verifikasi
-- (berpengaruh ke posting jurnal otomatis). Karena itu kolom kembali nullable.
-- Lihat docs/plan/payment-manual-santri.md.

ALTER TABLE payments ALTER COLUMN debit_account_id DROP NOT NULL;
