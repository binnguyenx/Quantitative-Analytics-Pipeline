package service

import (
	"context"
	"errors"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/finbud/finbud-backend/internal/domain"
)

type fakeIngestRepo struct {
	createErr     error
	createCalls   int
	storedBatches int
}

func (f *fakeIngestRepo) CreateBatch(_ context.Context, events []domain.IngestedEvent) (int64, error) {
	f.createCalls++
	if f.createErr != nil {
		return 0, f.createErr
	}
	f.storedBatches += len(events)
	return int64(len(events)), nil
}

type fakeDLQPublisher struct {
	publishErr   error
	publishCalls int
}

func (f *fakeDLQPublisher) Publish(_ context.Context, _, _ []byte) error {
	f.publishCalls++
	return f.publishErr
}

func TestIngestionService_RoutesInvalidToDLQ(t *testing.T) {
	repo := &fakeIngestRepo{}
	dlq := &fakeDLQPublisher{}
	svc := NewIngestionService(repo, nil, dlq, 1, time.Millisecond)

	msgs := []kafkago.Message{
		{
			Topic: "EVENT_PROFILE_UPDATED",
			Key:   []byte("k1"),
			Value: []byte(`{"event":"EVENT_PROFILE_UPDATED","user_id":"bad","timestamp":"2026-01-01T10:00:00Z"}`),
		},
	}

	if err := svc.ProcessBatch(context.Background(), msgs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dlq.publishCalls != 1 {
		t.Fatalf("expected 1 dlq publish call, got %d", dlq.publishCalls)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected no db create for invalid payload, got %d", repo.createCalls)
	}
}

func TestIngestionService_RetryOnPersistFailure(t *testing.T) {
	repo := &fakeIngestRepo{createErr: errors.New("db fail")}
	dlq := &fakeDLQPublisher{}
	svc := NewIngestionService(repo, nil, dlq, 2, time.Millisecond)

	msgs := []kafkago.Message{
		{
			Topic: "EVENT_PROFILE_UPDATED",
			Key:   []byte("k2"),
			Value: []byte(`{
				"event":"EVENT_PROFILE_UPDATED",
				"event_id":"evt-1",
				"user_id":"8f92de77-13d5-4796-ae0e-2f4ed748ff7f",
				"timestamp":"2026-01-01T10:00:00Z",
				"data":{"x":1}
			}`),
		},
	}

	err := svc.ProcessBatch(context.Background(), msgs)
	if err == nil {
		t.Fatal("expected persist error")
	}
	if repo.createCalls != 3 {
		t.Fatalf("expected 3 create attempts, got %d", repo.createCalls)
	}
}
