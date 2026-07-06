CREATE TABLE IF NOT EXISTS transfers (
    id VARCHAR(64) PRIMARY KEY,
    external_id VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    amount_cents BIGINT NOT NULL,
    payee_name VARCHAR(255) NOT NULL,
    payee_tax_id VARCHAR(32) NOT NULL,
    payee_pix_key VARCHAR(128) NOT NULL,
    payee_pix_type VARCHAR(16),
    payee_bank_code VARCHAR(16),
    payee_branch VARCHAR(16),
    payee_account VARCHAR(32),
    payee_account_type VARCHAR(16),
    provider VARCHAR(32) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    metadata JSONB DEFAULT '{}',
    processed_at TIMESTAMPTZ,
    settled_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    failure_reason TEXT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
CREATE INDEX IF NOT EXISTS idx_transfers_external_id ON transfers(external_id);
CREATE INDEX IF NOT EXISTS idx_transfers_payee_tax_id ON transfers(payee_tax_id);