CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE
    IF NOT EXISTS accounts (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        name TEXT NOT NULL,
        bank_id TEXT NOT NULL,
        balance NUMERIC(18, 2) DEFAULT 0,
        created_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT now ()
    );

CREATE INDEX IF NOT EXISTS idx_accounts_bank_id ON accounts (bank_id);