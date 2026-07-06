CREATE TABLE IF NOT EXISTS webhooks (
    id VARCHAR(64) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    events TEXT[] NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    secret VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhooks_client_id ON webhooks(client_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(active);