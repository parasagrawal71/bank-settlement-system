CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE
    IF NOT EXISTS accounts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        name TEXT NOT NULL,
        account_no TEXT NOT NULL,
        balance DOUBLE PRECISION NOT NULL DEFAULT 0,
        reserved DOUBLE PRECISION NOT NULL DEFAULT 0,
        created_at TIMESTAMP DEFAULT NOW (),
        updated_at TIMESTAMP DEFAULT NOW ()
    );

CREATE INDEX IF NOT EXISTS idx_accounts_account_id ON accounts (account_no);

CREATE TYPE reservation_status_enum AS ENUM ('PENDING', 'CONFIRMED', 'FAILED');

CREATE TABLE
    IF NOT EXISTS reservations (
        reference_id VARCHAR(100) PRIMARY KEY,
        payer_id VARCHAR(64) NOT NULL,
        payee_id VARCHAR(64) NOT NULL,
        amount DOUBLE PRECISION NOT NULL,
        status reservation_status_enum DEFAULT 'PENDING',
        created_at TIMESTAMP DEFAULT NOW (),
        updated_at TIMESTAMP DEFAULT NOW ()
    );

CREATE TABLE ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payer_id UUID,
    payee_id UUID,
    amount NUMERIC(12,2) NOT NULL,
    reference_id VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(20) CHECK (status IN ('INITIATED', 'COMPLETED', 'FAILED')) NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);
