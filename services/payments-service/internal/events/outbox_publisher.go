package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/parasagrawal71/bank-settlement-system/services/payments-service/internal/repository"
)

type OutboxPublisher struct {
	pool     *pgxpool.Pool
	repo     *repository.OutboxRepository
	producer *Producer
}

func NewOutboxPublisher(pool *pgxpool.Pool, repo *repository.OutboxRepository, producer *Producer) *OutboxPublisher {
	return &OutboxPublisher{pool: pool, repo: repo, producer: producer}
}

func (p *OutboxPublisher) Start(ctx context.Context) {
	log.Println("OutboxPublisher started")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("OutboxPublisher stopped")
			return
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

func (p *OutboxPublisher) processBatch(ctx context.Context) {
	log.Println("OutboxPublisher processing batch")
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Printf("outbox begin tx: %v", err)
		return
	}
	defer tx.Rollback(ctx)

	events, err := p.repo.FetchPending(ctx, tx)
	if err != nil {
		log.Printf("fetch pending outbox: %v", err)
		return
	}

	// nothing to do
	if len(events) == 0 {
		log.Println("no pending outbox events")
		return
	}

	for _, e := range events {
		var ev PaymentEvent
		if err := json.Unmarshal(e.Payload, &ev); err != nil {
			log.Printf("invalid outbox payload: %v. Marking as failed", err)
			_ = p.repo.MarkAsFailed(ctx, tx, e.ID)
			continue
		}

		// check retry count
		if e.RetryCount >= 3 {
			log.Printf("outbox retry limit reached: %v. Marking as failed", err)
			_ = p.repo.MarkAsFailed(ctx, tx, e.ID)
			continue
		}

		if err := p.producer.PublishEvent(ctx, ev); err != nil {
			log.Printf("publish fail: %v. Incrementing retry count", err)
			_ = p.repo.IncrementRetry(ctx, tx, e.ID)
			continue
		}

		log.Printf("published event: %v", ev)
		_ = p.repo.MarkAsPublished(ctx, tx, e.ID)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("outbox commit err: %v", err)
	}
}
