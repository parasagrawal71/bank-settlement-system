package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	topic  string
}

type PaymentEvent struct {
	ReferenceID string  `json:"reference_id"`
	AccountID   string  `json:"account_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	TxnType     string  `json:"txn_type"` // DEBIT or CREDIT
	Timestamp   int64   `json:"timestamp"`
	// optional: correlation ids, source, environment
}

func NewProducer(brokers []string, topic string) *Producer {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Async:    false, // we will write synchronously in our wrapper
	})
	return &Producer{writer: w, topic: topic}
}

func (p *Producer) Close(ctx context.Context) error {
	return p.writer.Close()
}

// PublishEvent attempts to write event with retries and timeout.
// Returns error only after exhausting retries.
func (p *Producer) PublishEvent(ctx context.Context, ev PaymentEvent) error {
	msgBytes, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// short timeout for each attempt
	attempts := 3
	var lastErr error
	for i := 0; i < attempts; i++ {
		ctxWrite, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = p.writer.WriteMessages(ctxWrite, kafka.Message{
			Key:   []byte(ev.ReferenceID), // partition by reference for ordering
			Value: msgBytes,
		})
		cancel()
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(time.Duration(200*(i+1)) * time.Millisecond) // simple backoff
	}
	return fmt.Errorf("publish failed after %d attempts: %w", attempts, lastErr)
}

func EnsureTopicExists(broker string, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	ctrlConn, err := kafka.Dial("tcp", controller.Host+":"+strconv.Itoa(controller.Port))
	if err != nil {
		return err
	}
	defer ctrlConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
	}
	return ctrlConn.CreateTopics(topicConfigs...)
}
