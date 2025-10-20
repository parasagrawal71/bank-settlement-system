package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Settlement struct {
	PayerID     string
	PayeeID     string
	Amount      float64
	ReferenceID string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SettlementRepository struct {
	pool *pgxpool.Pool
}

func NewSettlementRepository(pool *pgxpool.Pool) *SettlementRepository {
	return &SettlementRepository{pool: pool}
}

func (r *SettlementRepository) CreateOrUpdate(ctx context.Context, s Settlement) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO settlements (payer_id, payee_id, amount, reference_id, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (reference_id) DO UPDATE SET status = EXCLUDED.status, updated_at = now()
	`, s.PayerID, s.PayeeID, s.Amount, s.ReferenceID, s.Status)
	return err
}

func (r *SettlementRepository) GetByReferenceID(ctx context.Context, ref string) (*Settlement, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT payer_id, payee_id, amount, reference_id, status, created_at, updated_at
		FROM settlements WHERE reference_id=$1
	`, ref)

	var s Settlement
	err := row.Scan(&s.PayerID, &s.PayeeID, &s.Amount, &s.ReferenceID, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
