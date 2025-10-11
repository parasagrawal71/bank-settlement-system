CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    account_id VARCHAR(50) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    txn_type VARCHAR(10) CHECK (txn_type IN ('DEBIT', 'CREDIT')) NOT NULL,
    status VARCHAR(20) DEFAULT 'SUCCESS',
    reference_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create a unique index on (reference_id, txn_type)
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_reference_txn_type
    ON payments (reference_id, txn_type);