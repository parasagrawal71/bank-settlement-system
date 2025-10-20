CREATE TABLE IF NOT EXISTS settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payer_id UUID,
    payee_id UUID,
    amount NUMERIC(12,2) NOT NULL,
    reference_id VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(20) CHECK (status IN ('PENDING', 'SETTLED', 'FAILED')) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_settlement_reference_id ON settlements(reference_id);
