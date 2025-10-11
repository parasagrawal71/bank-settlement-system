package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxEvent struct {
	ID         int
	EventType  string
	Payload    []byte
	Status     string
	RetryCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

func (r *OutboxRepository) AddEvent(ctx context.Context, tx pgx.Tx, eventType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (event_type, payload)
		VALUES ($1, $2)
	`, eventType, data)
	return err
}

func (r *OutboxRepository) FetchPending(ctx context.Context, conn pgx.Tx) ([]OutboxEvent, error) {
	rows, err := conn.Query(ctx, `
		SELECT id, event_type, payload, retry_count
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.RetryCount); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *OutboxRepository) MarkAsPublished(ctx context.Context, conn pgx.Tx, id int) error {
	_, err := conn.Exec(ctx, `
		UPDATE outbox_events
		SET status = 'PUBLISHED', updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (r *OutboxRepository) IncrementRetry(ctx context.Context, conn pgx.Tx, id int) error {
	_, err := conn.Exec(ctx, `
		UPDATE outbox_events
		SET retry_count = retry_count + 1, updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (r *OutboxRepository) MarkAsFailed(ctx context.Context, conn pgx.Tx, id int) error {
	_, err := conn.Exec(ctx, `
		UPDATE outbox_events
		SET status = 'FAILED', updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}
