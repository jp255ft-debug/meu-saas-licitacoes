CREATE TABLE IF NOT EXISTS payments (
    id              VARCHAR(64) PRIMARY KEY,
    external_id     VARCHAR(128),
    status          VARCHAR(32) NOT NULL,
    payment_type    VARCHAR(32) NOT NULL,
    method          VARCHAR(32) NOT NULL,
    amount_cents    BIGINT NOT NULL,
    fee_cents       BIGINT NOT NULL DEFAULT 0,
    net_amount_cents BIGINT NOT NULL,
    payer_name      VARCHAR(255),
    payer_tax_id    VARCHAR(32) NOT NULL,
    payer_email     VARCHAR(255),
    payer_pix_key   VARCHAR(128),
    payer_pix_type  VARCHAR(16),
    payee_name      VARCHAR(255),
    payee_tax_id    VARCHAR(32) NOT NULL,
    payee_pix_key   VARCHAR(128),
    payee_pix_type  VARCHAR(16),
    payee_bank_code VARCHAR(16),
    payee_branch    VARCHAR(16),
    payee_account   VARCHAR(32),
    payee_account_type VARCHAR(16),
    description     TEXT,
    expires_at      TIMESTAMPTZ,
    processed_at    TIMESTAMPTZ,
    settled_at      TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ,
    failure_reason  TEXT,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version         INT NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_external_id ON payments(external_id);
CREATE INDEX IF NOT EXISTS idx_payments_payer_tax_id ON payments(payer_tax_id);
CREATE INDEX IF NOT EXISTS idx_payments_expires_at ON payments(expires_at) WHERE status = 'PENDING';
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);
