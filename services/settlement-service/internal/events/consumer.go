package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/parasagrawal71/bank-settlement-system/services/settlement-service/internal/repository"
	"github.com/segmentio/kafka-go"
)

type PaymentCapturedEvent struct {
	ReferenceID string  `json:"reference_id"`
	PayerID     string  `json:"payer_id"`
	PayeeID     string  `json:"payee_id"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
}

type Consumer struct {
	reader *kafka.Reader
	repo   *repository.SettlementRepository
}

func NewConsumer(broker, topic, groupID string, pool *pgxpool.Pool) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		GroupID: groupID,
		Topic:   topic,
	})
	return &Consumer{reader: r, repo: repository.NewSettlementRepository(pool)}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Settlement consumer started for topic: payments_events")

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("consumer fetch error: %v", err)
			continue
		}

		var ev PaymentCapturedEvent
		if err := json.Unmarshal(msg.Value, &ev); err != nil {
			log.Printf("invalid event payload: %v", err)
			continue
		}

		settlement := repository.Settlement{
			ReferenceID: ev.ReferenceID,
			PayerID:     ev.PayerID,
			PayeeID:     ev.PayeeID,
			Amount:      ev.Amount,
			Status:      "PENDING",
		}

		if err := c.repo.CreateOrUpdate(ctx, settlement); err != nil {
			log.Printf("failed to create settlement: %v", err)
			continue
		}

		// mark message as processed
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("failed to commit message: %v", err)
		}
	}
}
