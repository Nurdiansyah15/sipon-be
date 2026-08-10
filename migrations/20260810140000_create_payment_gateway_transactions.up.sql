-- ============================================================
-- MIDTRANS PAYMENT GATEWAY TRANSACTIONS
-- ============================================================

CREATE TABLE IF NOT EXISTS payment_gateway_transactions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id    VARCHAR(50) NOT NULL UNIQUE,
    invoice_id        UUID NOT NULL REFERENCES invoices(id),
    payment_id        UUID REFERENCES payments(id),
    amount            NUMERIC(14,2) NOT NULL,
    status            VARCHAR(30) NOT NULL DEFAULT 'pending'
                      CHECK (status IN (
                          'pending', 'pending_challenge', 'capture',
                          'settlement', 'deny', 'failure', 'expire', 'cancel'
                      )),
    payment_method    VARCHAR(50),
    snap_token        TEXT NOT NULL,
    redirect_url      TEXT NOT NULL,
    raw_notification  JSONB,
    metadata          JSONB,
    expired_at        TIMESTAMPTZ NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_pg_tx_transaction_id ON payment_gateway_transactions(transaction_id);
CREATE INDEX idx_pg_tx_invoice ON payment_gateway_transactions(invoice_id);
CREATE INDEX idx_pg_tx_status ON payment_gateway_transactions(status);
CREATE INDEX idx_pg_tx_payment ON payment_gateway_transactions(payment_id);
CREATE INDEX idx_pg_tx_active_invoice ON payment_gateway_transactions(invoice_id, status)
    WHERE status IN ('pending', 'pending_challenge', 'capture', 'settlement');
