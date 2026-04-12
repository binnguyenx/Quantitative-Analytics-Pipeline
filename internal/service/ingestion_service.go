package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"

	"github.com/finbud/finbud-backend/internal/domain"
	"github.com/finbud/finbud-backend/internal/ingestion"
	"github.com/finbud/finbud-backend/internal/metrics"
	"github.com/finbud/finbud-backend/internal/repository"
)

type dlqPublisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

// IngestionService processes Kafka events and persists them idempotently.
type IngestionService struct {
	ingestRepo    repository.IngestedEventRepository
	rdb           *redis.Client
	dlqPublisher  dlqPublisher
	maxRetries    int
	retryBackoff  time.Duration
	processedStat string
}

func NewIngestionService(
	ingestRepo repository.IngestedEventRepository,
	rdb *redis.Client,
	dlqPublisher dlqPublisher,
	maxRetries int,
	retryBackoff time.Duration,
) *IngestionService {
	if maxRetries < 0 {
		maxRetries = 0
	}
	return &IngestionService{
		ingestRepo:    ingestRepo,
		rdb:           rdb,
		dlqPublisher:  dlqPublisher,
		maxRetries:    maxRetries,
		retryBackoff:  retryBackoff,
		processedStat: "ingest:processed_total",
	}
}

// ProcessBatch validates, stores and routes failed messages to DLQ.
func (s *IngestionService) ProcessBatch(ctx context.Context, msgs []kafkago.Message) error {
	if len(msgs) == 0 {
		return nil
	}
	metrics.ObserveIngestBatch(len(msgs))

	records := make([]domain.IngestedEvent, 0, len(msgs))
	startedAt := time.Now()

	for _, msg := range msgs {
		eventStart := time.Now()
		if msg.HighWaterMark > 0 && msg.Offset >= 0 {
			metrics.SetConsumerLag(msg.HighWaterMark - msg.Offset)
		}

		record, err := ingestion.ParseProfileUpdatedMessage(msg)
		if err != nil {
			metrics.ObserveIngestEvent("failed", float64(time.Since(eventStart).Milliseconds()))
			if dlqErr := s.publishDLQ(ctx, msg, err); dlqErr != nil {
				return fmt.Errorf("process dlq publish: %w", dlqErr)
			}
			metrics.ObserveIngestEvent("dlq", float64(time.Since(eventStart).Milliseconds()))
			s.logJSON("warn", "invalid message routed to dlq", map[string]interface{}{
				"error":  err.Error(),
				"topic":  msg.Topic,
				"offset": msg.Offset,
			})
			continue
		}

		records = append(records, record)
	}

	if len(records) > 0 {
		_, err := s.withRetry(ctx, func(execCtx context.Context) error {
			_, createErr := s.ingestRepo.CreateBatch(execCtx, records)
			return createErr
		})
		if err != nil {
			return fmt.Errorf("persist batch: %w", err)
		}

		for range records {
			metrics.ObserveIngestEvent("success", float64(time.Since(startedAt).Milliseconds()))
		}

		if s.rdb != nil {
			_ = s.rdb.IncrBy(ctx, s.processedStat, int64(len(records))).Err()
		}
	}

	s.logJSON("info", "ingestion batch processed", map[string]interface{}{
		"batch_size":      len(msgs),
		"stored_records":  len(records),
		"duration_ms":     time.Since(startedAt).Milliseconds(),
		"consumer_status": "ok",
	})

	return nil
}

func (s *IngestionService) withRetry(ctx context.Context, fn func(context.Context) error) (int, error) {
	var err error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		err = fn(ctx)
		if err == nil {
			return attempt, nil
		}
		if attempt == s.maxRetries {
			break
		}
		sleep := s.retryBackoff * time.Duration(1<<attempt)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return attempt, ctx.Err()
		}
	}
	return s.maxRetries, err
}

func (s *IngestionService) publishDLQ(ctx context.Context, msg kafkago.Message, cause error) error {
	if s.dlqPublisher == nil {
		return fmt.Errorf("dlq publisher not configured: %w", cause)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"failed_at": time.Now().UTC().Format(time.RFC3339),
		"error":     cause.Error(),
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"key":       string(msg.Key),
		"value":     string(msg.Value),
	})

	_, err := s.withRetry(ctx, func(execCtx context.Context) error {
		return s.dlqPublisher.Publish(execCtx, msg.Key, payload)
	})
	return err
}

func (s *IngestionService) logJSON(level, message string, fields map[string]interface{}) {
	data := map[string]interface{}{
		"level":   level,
		"message": message,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range fields {
		data[k] = v
	}

	out, err := json.Marshal(data)
	if err != nil {
		log.Printf("ingestionService: marshal log error: %v", err)
		return
	}
	log.Println(string(out))
}
