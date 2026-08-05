-- ============================================================
-- BILLING TABLES
-- ============================================================

CREATE TABLE IF NOT EXISTS fee_components (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(20) NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    type        VARCHAR(30) NOT NULL CHECK (type IN ('ukt', 'spp', 'daftar_ulang', 'insidental')),
    amount      NUMERIC(14,2) NOT NULL DEFAULT 0,
    is_periodic BOOLEAN NOT NULL DEFAULT false,
    period_type VARCHAR(20) CHECK (period_type IN ('monthly', 'semesterly', 'yearly', 'once')),
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_fee_components_type ON fee_components(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_fee_components_active ON fee_components(is_active) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS billing_schemes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing_scheme_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    billing_scheme_id UUID NOT NULL REFERENCES billing_schemes(id) ON DELETE CASCADE,
    fee_component_id  UUID NOT NULL REFERENCES fee_components(id),
    amount_override   NUMERIC(14,2),
    is_required       BOOLEAN NOT NULL DEFAULT true,
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(billing_scheme_id, fee_component_id)
);
CREATE INDEX idx_bsi_scheme ON billing_scheme_items(billing_scheme_id);

-- No FK to santri — cross-module reference.
CREATE TABLE IF NOT EXISTS santri_billing_assignments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_id         UUID NOT NULL,
    billing_scheme_id UUID NOT NULL REFERENCES billing_schemes(id),
    effective_from    DATE NOT NULL,
    effective_until   DATE,
    assigned_by       UUID NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sba_santri ON santri_billing_assignments(santri_id);
CREATE INDEX idx_sba_active ON santri_billing_assignments(santri_id, effective_from);

-- No FK to santri/users — cross-module references.
CREATE TABLE IF NOT EXISTS invoices (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number    VARCHAR(30) NOT NULL UNIQUE,
    santri_id         UUID NOT NULL,
    user_id           UUID NOT NULL,
    billing_scheme_id UUID,
    fee_component_id  UUID NOT NULL REFERENCES fee_components(id),
    periode           VARCHAR(20) NOT NULL,
    tahun_ajaran      VARCHAR(10) NOT NULL,
    amount            NUMERIC(14,2) NOT NULL,
    discount_amount   NUMERIC(14,2) NOT NULL DEFAULT 0,
    paid_amount       NUMERIC(14,2) NOT NULL DEFAULT 0,
    status            VARCHAR(20) NOT NULL DEFAULT 'draft'
                      CHECK (status IN ('draft', 'issued', 'partial', 'paid', 'expired', 'cancelled')),
    due_date          DATE NOT NULL,
    issued_at         DATE,
    notes             TEXT,
    created_by        UUID NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_invoices_number ON invoices(invoice_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_santri ON invoices(santri_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_user ON invoices(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_status ON invoices(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_tahun_ajaran ON invoices(tahun_ajaran) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_periode ON invoices(periode) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_due_date ON invoices(due_date) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_invoices_unique_periode
    ON invoices(santri_id, fee_component_id, periode)
    WHERE deleted_at IS NULL AND status NOT IN ('cancelled');

CREATE TABLE IF NOT EXISTS payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_number    VARCHAR(30) NOT NULL UNIQUE,
    invoice_id        UUID NOT NULL REFERENCES invoices(id),
    debit_account_id  UUID,
    amount            NUMERIC(14,2) NOT NULL,
    method            VARCHAR(20) NOT NULL CHECK (method IN ('transfer', 'cash', 'check')),
    reference_number  VARCHAR(100),
    payment_date      DATE NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending', 'verified', 'rejected')),
    verified_by       UUID,
    verified_at       TIMESTAMPTZ,
    notes             TEXT,
    proof_key         VARCHAR(512),
    created_by        UUID NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_payments_number ON payments(payment_number);
CREATE INDEX idx_payments_invoice ON payments(invoice_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_date ON payments(payment_date);

CREATE TABLE IF NOT EXISTS invoice_adjustments (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id    UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    type          VARCHAR(20) NOT NULL CHECK (type IN ('beasiswa', 'diskon', 'penyesuaian')),
    amount        NUMERIC(14,2) NOT NULL DEFAULT 0,
    percentage    NUMERIC(5,2),
    description   TEXT,
    applied_by    UUID NOT NULL,
    applied_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_adjustments_invoice ON invoice_adjustments(invoice_id);

-- ============================================================
-- ACCOUNTING TABLES
-- ============================================================

CREATE TABLE IF NOT EXISTS accounts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code           VARCHAR(20) NOT NULL UNIQUE,
    name           VARCHAR(200) NOT NULL,
    type           VARCHAR(20) NOT NULL CHECK (type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
    parent_id      UUID REFERENCES accounts(id),
    level          INTEGER NOT NULL DEFAULT 0,
    is_postable    BOOLEAN NOT NULL DEFAULT false,
    normal_balance VARCHAR(10) NOT NULL CHECK (normal_balance IN ('debit', 'credit')),
    description    TEXT,
    is_active      BOOLEAN NOT NULL DEFAULT true,
    is_system      BOOLEAN NOT NULL DEFAULT false,
    created_by     UUID NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);
CREATE INDEX idx_accounts_type ON accounts(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_parent ON accounts(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_postable ON accounts(is_postable) WHERE deleted_at IS NULL AND is_active = true;

CREATE TABLE IF NOT EXISTS accounting_periods (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'open'
               CHECK (status IN ('open', 'closing', 'closed', 'locked')),
    closed_by  UUID,
    closed_at  TIMESTAMPTZ,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_periods_status ON accounting_periods(status);
CREATE INDEX idx_periods_date_range ON accounting_periods(start_date, end_date);

CREATE OR REPLACE FUNCTION check_period_overlap()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM accounting_periods
        WHERE id != NEW.id
          AND start_date <= NEW.end_date
          AND end_date >= NEW.start_date
    ) THEN
        RAISE EXCEPTION 'period overlaps with existing period';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_period_overlap
    BEFORE INSERT OR UPDATE ON accounting_periods
    FOR EACH ROW
    EXECUTE FUNCTION check_period_overlap();

CREATE TABLE IF NOT EXISTS journal_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_number  VARCHAR(30) NOT NULL UNIQUE,
    entry_date      DATE NOT NULL,
    description     TEXT NOT NULL,
    source_type     VARCHAR(30) CHECK (source_type IN (
                        'invoice_issued', 'payment_verified', 'invoice_cancelled',
                        'adjustment', 'closing', 'manual'
                    )),
    source_id       UUID,
    period_id       UUID NOT NULL REFERENCES accounting_periods(id),
    total_debit     NUMERIC(16,2) NOT NULL,
    total_credit    NUMERIC(16,2) NOT NULL,
    posted_by       UUID NOT NULL,
    posted_at       TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'posted', 'cancelled')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_journal_number ON journal_entries(journal_number);
CREATE INDEX idx_journal_date ON journal_entries(entry_date);
CREATE INDEX idx_journal_period ON journal_entries(period_id);
CREATE INDEX idx_journal_source ON journal_entries(source_type, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX idx_journal_status ON journal_entries(status);
ALTER TABLE journal_entries ADD CONSTRAINT chk_journal_balance
    CHECK (total_debit = total_credit);

CREATE TABLE IF NOT EXISTS journal_entry_lines (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id       UUID NOT NULL REFERENCES accounts(id),
    account_code     VARCHAR(20) NOT NULL,
    description      TEXT,
    debit            NUMERIC(16,2) NOT NULL DEFAULT 0,
    credit           NUMERIC(16,2) NOT NULL DEFAULT 0
);
CREATE INDEX idx_journal_lines_entry ON journal_entry_lines(journal_entry_id);
CREATE INDEX idx_journal_lines_account ON journal_entry_lines(account_id);
ALTER TABLE journal_entry_lines ADD CONSTRAINT chk_line_debit_xor_credit
    CHECK ((debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0));

-- ============================================================
-- AUDIT
-- ============================================================

CREATE TABLE IF NOT EXISTS finance_audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id   UUID NOT NULL,
    action      VARCHAR(50) NOT NULL,
    actor_id    UUID NOT NULL,
    changes     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_fal_entity ON finance_audit_logs(entity_type, entity_id);
CREATE INDEX idx_fal_actor ON finance_audit_logs(actor_id);
