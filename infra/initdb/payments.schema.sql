-- payment_intents table
CREATE TABLE payment_intents (
  id SERIAL PRIMARY KEY,
  reference_id VARCHAR(100) UNIQUE NOT NULL,
  payer_id VARCHAR(100) NOT NULL,
  payee_id VARCHAR(100) NOT NULL,
  amount NUMERIC(12,2) NOT NULL,
  status VARCHAR(20) CHECK (status IN ('AUTHORIZED', 'CAPTURED', 'FAILED')) NOT NULL,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);


-- payments table (capture creates two rows with same reference_id)
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    account_id VARCHAR(50) NOT NULL,
    amount NUMERIC(12, 2) NOT NULL,
    txn_type VARCHAR(10) CHECK (txn_type IN ('DEBIT', 'CREDIT')) NOT NULL,
    status VARCHAR(20) DEFAULT 'SUCCESS',
    reference_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create a unique index on (reference_id, txn_type)
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_reference_txn_type
    ON payments (reference_id, txn_type);


-- outbox events table
CREATE TABLE IF NOT EXISTS outbox_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- idempotency keys (store final response)
CREATE TABLE idempotency_keys (
  key VARCHAR(100) PRIMARY KEY,
  created_at TIMESTAMP DEFAULT now(),
  response JSONB
);
