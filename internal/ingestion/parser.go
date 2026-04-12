package ingestion

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"

	"github.com/finbud/finbud-backend/internal/domain"
)

const EventProfileUpdated = "EVENT_PROFILE_UPDATED"

type profileUpdatedEnvelope struct {
	Event         string          `json:"event"`
	EventID       string          `json:"event_id"`
	CorrelationID string          `json:"correlation_id"`
	UserID        string          `json:"user_id"`
	Timestamp     string          `json:"timestamp"`
	Data          json.RawMessage `json:"data"`
}

// ParseProfileUpdatedMessage validates payload schema and maps a Kafka message
// to the ingested-event domain model.
func ParseProfileUpdatedMessage(msg kafkago.Message) (domain.IngestedEvent, error) {
	var envelope profileUpdatedEnvelope
	if err := json.Unmarshal(msg.Value, &envelope); err != nil {
		return domain.IngestedEvent{}, fmt.Errorf("parse payload: %w", err)
	}

	if envelope.Event != EventProfileUpdated {
		return domain.IngestedEvent{}, fmt.Errorf("unexpected event type: %s", envelope.Event)
	}

	userID, err := uuid.Parse(envelope.UserID)
	if err != nil {
		return domain.IngestedEvent{}, fmt.Errorf("invalid user_id: %w", err)
	}

	eventTS, err := time.Parse(time.RFC3339, envelope.Timestamp)
	if err != nil {
		return domain.IngestedEvent{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	eventID := envelope.EventID
	if eventID == "" {
		eventID = deriveEventID(msg.Key, envelope.Timestamp, msg.Value)
	}

	correlationID := envelope.CorrelationID
	if correlationID == "" {
		correlationID = eventID
	}

	return domain.IngestedEvent{
		EventID:        eventID,
		EventType:      envelope.Event,
		UserID:         userID,
		CorrelationID:  correlationID,
		EventTimestamp: eventTS.UTC(),
		KafkaKey:       string(msg.Key),
		KafkaTopic:     msg.Topic,
		KafkaPartition: msg.Partition,
		KafkaOffset:    msg.Offset,
		PayloadJSON:    string(msg.Value),
		ProcessedAt:    time.Now().UTC(),
	}, nil
}

func deriveEventID(key []byte, timestamp string, payload []byte) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s", string(key), timestamp, string(payload))))
	return hex.EncodeToString(sum[:])
}
