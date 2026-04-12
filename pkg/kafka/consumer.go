package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// Consumer wraps kafka-go reader with batch fetch helpers.
type Consumer struct {
	reader *kafkago.Reader
}

func NewConsumer(brokers, topic, groupID string) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:         splitBrokers(brokers),
		GroupID:         groupID,
		Topic:           topic,
		MinBytes:        1,
		MaxBytes:        10e6,
		CommitInterval:  0, // manual commit to keep batch control
		ReadLagInterval: -1,
	})
	return &Consumer{reader: reader}
}

func (c *Consumer) FetchBatch(ctx context.Context, batchSize int, maxWait time.Duration) ([]kafkago.Message, error) {
	if batchSize <= 0 {
		return nil, fmt.Errorf("batch size must be positive")
	}
	if maxWait <= 0 {
		maxWait = 500 * time.Millisecond
	}

	deadline := time.Now().Add(maxWait)
	msgs := make([]kafkago.Message, 0, batchSize)

	for len(msgs) < batchSize {
		if time.Now().After(deadline) {
			break
		}

		fetchCtx, cancel := context.WithDeadline(ctx, deadline)
		msg, err := c.reader.FetchMessage(fetchCtx)
		cancel()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				break
			}
			return nil, fmt.Errorf("fetch message: %w", err)
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (c *Consumer) CommitBatch(ctx context.Context, msgs []kafkago.Message) error {
	if len(msgs) == 0 {
		return nil
	}
	if err := c.reader.CommitMessages(ctx, msgs...); err != nil {
		return fmt.Errorf("commit messages: %w", err)
	}
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func splitBrokers(brokers string) []string {
	return splitCSV(brokers)
}
