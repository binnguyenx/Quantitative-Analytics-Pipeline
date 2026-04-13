package telemetry

import (
	"encoding/json"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type metricEventPayload struct {
	Event      string                 `json:"event"`
	Timestamp  string                 `json:"timestamp"`
	MetricName string                 `json:"metric_name"`
	Value      float64                `json:"value"`
	Source     string                 `json:"source"`
	Tags       map[string]interface{} `json:"tags"`
	Data       map[string]interface{} `json:"data"`
}

func ParseKafkaMessage(msg kafkago.Message) ([]NormalizedEvent, error) {
	var payload metricEventPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return nil, fmt.Errorf("decode kafka payload: %w", err)
	}
	ts := parseTimestamp(payload.Timestamp)
	source := payload.Source
	if source == "" {
		source = "kafka"
	}

	events := make([]NormalizedEvent, 0, 4)
	if payload.MetricName != "" {
		events = append(events, NormalizedEvent{
			Timestamp:  ts,
			Source:     source,
			MetricName: payload.MetricName,
			Value:      payload.Value,
			Tags:       stringifyMap(payload.Tags),
		})
	}

	// Generic event count for every consumed event.
	baseTags := map[string]string{
		"event_type": payload.Event,
		"status":     statusFromPayload(payload),
	}
	events = append(events, NormalizedEvent{
		Timestamp:  ts,
		Source:     source,
		MetricName: "event_processed",
		Value:      1,
		Tags:       baseTags,
	})

	// Optional latency metric from envelope data.
	if latency, ok := getFloat(payload.Data, "latency_ms"); ok {
		events = append(events, NormalizedEvent{
			Timestamp:  ts,
			Source:     source,
			MetricName: "latency_ms",
			Value:      latency,
			Tags:       baseTags,
		})
	}
	if lag, ok := getFloat(payload.Data, "consumer_lag"); ok {
		events = append(events, NormalizedEvent{
			Timestamp:  ts,
			Source:     source,
			MetricName: "consumer_lag",
			Value:      lag,
			Tags:       baseTags,
		})
	}

	return events, nil
}

func ParseRedisMessage(payload string) ([]NormalizedEvent, error) {
	var msg metricEventPayload
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		return nil, fmt.Errorf("decode redis payload: %w", err)
	}
	ts := parseTimestamp(msg.Timestamp)
	source := msg.Source
	if source == "" {
		source = "redis"
	}
	return []NormalizedEvent{
		{
			Timestamp:  ts,
			Source:     source,
			MetricName: msg.MetricName,
			Value:      msg.Value,
			Tags:       stringifyMap(msg.Tags),
		},
	}, nil
}

func parseTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Now().UTC()
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Now().UTC()
	}
	return t.UTC()
}

func statusFromPayload(payload metricEventPayload) string {
	if payload.Data == nil {
		return "ok"
	}
	if raw, ok := payload.Data["status"]; ok {
		if s, ok := raw.(string); ok && s != "" {
			return s
		}
	}
	if raw, ok := payload.Data["is_error"]; ok {
		if b, ok := raw.(bool); ok && b {
			return "error"
		}
	}
	return "ok"
}

func stringifyMap(src map[string]interface{}) map[string]string {
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

func getFloat(src map[string]interface{}, key string) (float64, bool) {
	if src == nil {
		return 0, false
	}
	raw, ok := src[key]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}
