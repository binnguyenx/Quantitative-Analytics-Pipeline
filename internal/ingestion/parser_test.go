package ingestion

import (
	"testing"

	kafkago "github.com/segmentio/kafka-go"
)

func TestParseProfileUpdatedMessage_WithEventID(t *testing.T) {
	msg := kafkago.Message{
		Topic:     "EVENT_PROFILE_UPDATED",
		Partition: 1,
		Offset:    42,
		Key:       []byte("key-1"),
		Value: []byte(`{
			"event":"EVENT_PROFILE_UPDATED",
			"event_id":"evt-123",
			"correlation_id":"corr-123",
			"user_id":"8f92de77-13d5-4796-ae0e-2f4ed748ff7f",
			"timestamp":"2026-01-01T10:00:00Z",
			"data":{"monthly_income":5000}
		}`),
	}

	record, err := ParseProfileUpdatedMessage(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.EventID != "evt-123" {
		t.Fatalf("unexpected event id: %s", record.EventID)
	}
	if record.CorrelationID != "corr-123" {
		t.Fatalf("unexpected correlation id: %s", record.CorrelationID)
	}
}

func TestParseProfileUpdatedMessage_DeriveEventIDWhenMissing(t *testing.T) {
	msg := kafkago.Message{
		Topic: "EVENT_PROFILE_UPDATED",
		Key:   []byte("key-2"),
		Value: []byte(`{
			"event":"EVENT_PROFILE_UPDATED",
			"user_id":"8f92de77-13d5-4796-ae0e-2f4ed748ff7f",
			"timestamp":"2026-01-01T10:00:00Z",
			"data":{"monthly_income":5000}
		}`),
	}

	record, err := ParseProfileUpdatedMessage(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if record.EventID == "" {
		t.Fatal("expected derived event id")
	}
	if record.CorrelationID != record.EventID {
		t.Fatal("expected correlation_id fallback to event_id")
	}
}

func TestParseProfileUpdatedMessage_InvalidPayload(t *testing.T) {
	msg := kafkago.Message{
		Value: []byte(`{"event":"EVENT_PROFILE_UPDATED","user_id":"invalid"}`),
	}

	_, err := ParseProfileUpdatedMessage(msg)
	if err == nil {
		t.Fatal("expected parse validation error")
	}
}
