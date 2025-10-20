package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyRepo struct {
	pool *pgxpool.Pool
}

func NewIdempotencyRepository(pool *pgxpool.Pool) *IdempotencyRepo {
	return &IdempotencyRepo{pool: pool}
}

func (i *IdempotencyRepo) GetResponse(ctx context.Context, key string) ([]byte, error) {
	var b []byte
	err := i.pool.QueryRow(ctx, `SELECT response FROM idempotency_keys WHERE key=$1`, key).Scan(&b)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func (i *IdempotencyRepo) SaveResponse(ctx context.Context, key string, resp []byte) error {
	_, err := i.pool.Exec(ctx, `INSERT INTO idempotency_keys (key, response, created_at) VALUES ($1,$2,now())`, key, resp)
	return err
}
