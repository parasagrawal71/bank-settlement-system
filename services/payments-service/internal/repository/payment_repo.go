package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

type PaymentIntent struct {
	ID          string
	ReferenceID string
	PayerID     string
	PayeeID     string
	Amount      float64
	Status      string
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *Repository) CreateIntent(ctx context.Context, referenceID string, payerID string, payeeID string, amount float64) error {
	_, err := r.pool.Exec(ctx, `
    INSERT INTO payment_intents (reference_id, payer_id, payee_id, amount, status, created_at)
    VALUES ($1,$2,$3,$4,'AUTHORIZED', now())
    `, referenceID, payerID, payeeID, amount)
	return err
}

func (r *Repository) GetIntent(ctx context.Context, referenceID string) (*PaymentIntent, error) {
	var pi PaymentIntent
	err := r.pool.QueryRow(ctx, `
	SELECT id, reference_id, payer_id, payee_id, amount, status FROM payment_intents WHERE reference_id=$1
	`, referenceID).Scan(&pi.ID, &pi.ReferenceID, &pi.PayerID, &pi.PayeeID, &pi.Amount, &pi.Status)
	return &pi, err
}

func (r *Repository) UpdateIntentStatusTx(ctx context.Context, tx pgx.Tx, referenceID string, status string) error {
	_, err := tx.Exec(ctx, `
	UPDATE payment_intents SET status=$1 WHERE reference_id=$2
	`, status, referenceID)
	return err
}

func (r *Repository) InsertPaymentTx(ctx context.Context, tx pgx.Tx, referenceID string, accountID string, txnType string, amount float64) error {
	_, err := tx.Exec(ctx, `
    INSERT INTO payments (reference_id, account_id, amount, txn_type, created_at)
    VALUES ($1,$2,$3,$4, now())
    `, referenceID, accountID, amount, txnType)
	return err
}
