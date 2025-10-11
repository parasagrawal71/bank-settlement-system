package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *Repository) InsertTransaction(ctx context.Context, tx pgx.Tx, accountID, currency, txnType, referenceID string, amount float64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO transactions (account_id, amount, currency, txn_type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, accountID, amount, currency, txnType, referenceID, time.Now())
	return err
}
