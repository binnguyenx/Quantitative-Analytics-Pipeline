// Package kafka provides a Kafka producer for the event-driven
// parts of FinBud.  When a user updates their financial profile
// we emit an EVENT_PROFILE_UPDATED message so downstream consumers
// (analytics, notification service, etc.) can react asynchronously.
package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/finbud/finbud-backend/internal/metrics"
)

// Producer wraps a kafka-go Writer with FinBud-specific helpers.
type Producer struct {
	writer *kafkago.Writer
}

// NewProducer creates a batched, async-safe Kafka writer.
// `brokers` is a comma-separated list (e.g. "broker1:9092,broker2:9092").
func NewProducer(brokers, topic string) *Producer {
	w := &kafkago.Writer{
		Addr:         kafkago.TCP(splitCSV(brokers)...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond, // low latency
		RequiredAcks: kafkago.RequireOne,
		Async:        false, // set true for fire-and-forget
	}
	log.Printf("kafka: producer ready → topic=%s brokers=%s ✓\n", topic, brokers)
	return &Producer{writer: w}
}

// Publish sends a single keyed message.  The key is typically
// the user ID so all events for one user land on the same partition.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	start := time.Now()
	msg := kafkago.Message{
		Key:   key,
		Value: value,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		metrics.ObserveKafkaPublish("failed", float64(time.Since(start).Milliseconds()))
		return fmt.Errorf("kafka: publish failed: %w", err)
	}
	metrics.ObserveKafkaPublish("success", float64(time.Since(start).Milliseconds()))
	return nil
}

// Close flushes pending messages and releases resources.
func (p *Producer) Close() error {
	return p.writer.Close()
}

