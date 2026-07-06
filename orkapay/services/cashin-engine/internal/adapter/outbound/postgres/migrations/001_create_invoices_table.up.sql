CREATE TABLE IF NOT EXISTS invoices (
    id VARCHAR(64) PRIMARY KEY,
    external_id VARCHAR(128),
    invoice_number VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    amount_cents BIGINT NOT NULL,
    due_date TIMESTAMPTZ NOT NULL,
    payer_name VARCHAR(255) NOT NULL,
    payer_tax_id VARCHAR(32) NOT NULL,
    payer_email VARCHAR(255),
    payer_phone VARCHAR(32),
    payer_address_line1 VARCHAR(255),
    payer_address_line2 VARCHAR(255),
    payer_city VARCHAR(128),
    payer_state VARCHAR(32),
    payer_zip_code VARCHAR(16),
    provider VARCHAR(32) NOT NULL,
    pix_qr_code TEXT,
    pix_key VARCHAR(128),
    boleto_barcode VARCHAR(64),
    boleto_digitable VARCHAR(64),
    metadata JSONB DEFAULT '{}',
    paid_amount_cents BIGINT NOT NULL DEFAULT 0,
    paid_at TIMESTAMPTZ,
    payment_method VARCHAR(32),
    cancelled_at TIMESTAMPTZ,
    failure_reason TEXT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_external_id ON invoices(external_id);
CREATE INDEX IF NOT EXISTS idx_invoices_payer_tax_id ON invoices(payer_tax_id);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date) WHERE status IN ('PENDING', 'ISSUED');