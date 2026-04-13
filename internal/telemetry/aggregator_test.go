package telemetry

import (
	"testing"
	"time"
)

func TestAggregatorComputesWindowMetrics(t *testing.T) {
	agg := NewAggregator(10)
	now := time.Now().UTC()

	agg.AddEvent(NormalizedEvent{
		Timestamp:  now.Add(-20 * time.Second),
		Source:     "kafka",
		MetricName: "event_processed",
		Value:      1,
		Tags:       map[string]string{"status": "ok"},
	})
	agg.AddEvent(NormalizedEvent{
		Timestamp:  now.Add(-10 * time.Second),
		Source:     "kafka",
		MetricName: "latency_ms",
		Value:      120,
		Tags:       map[string]string{"status": "error"},
	})
	agg.AddEvent(NormalizedEvent{
		Timestamp:  now.Add(-8 * time.Second),
		Source:     "kafka",
		MetricName: "latency_ms",
		Value:      240,
		Tags:       map[string]string{"status": "ok"},
	})

	snap := agg.Snapshot(now)
	m := snap.Windows["1m"]
	if m.EventCount != 3 {
		t.Fatalf("expected event_count=3, got %d", m.EventCount)
	}
	if m.ErrorCount != 1 {
		t.Fatalf("expected error_count=1, got %d", m.ErrorCount)
	}
	if m.LatencyMS.P95 < 120 {
		t.Fatalf("expected p95 latency >= 120, got %.2f", m.LatencyMS.P95)
	}
	if m.Throughput <= 0 {
		t.Fatalf("expected throughput > 0, got %.2f", m.Throughput)
	}
}

func TestAggregatorPrunesOldEvents(t *testing.T) {
	agg := NewAggregator(5)
	now := time.Now().UTC()
	agg.AddEvent(NormalizedEvent{
		Timestamp:  now.Add(-20 * time.Minute),
		Source:     "kafka",
		MetricName: "event_processed",
		Value:      1,
	})
	agg.AddEvent(NormalizedEvent{
		Timestamp:  now.Add(-10 * time.Second),
		Source:     "kafka",
		MetricName: "event_processed",
		Value:      1,
	})

	snap := agg.Snapshot(now)
	if got := snap.Windows["15m"].EventCount; got != 1 {
		t.Fatalf("expected only recent event in 15m window, got %d", got)
	}
}
