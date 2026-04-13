package telemetry

import "time"

const SchemaVersion = "v1"

// NormalizedEvent is the unified event shape used by the telemetry service.
type NormalizedEvent struct {
	Timestamp  time.Time         `json:"timestamp"`
	Source     string            `json:"source"`
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// LatencyStats holds percentile metrics for latency data.
type LatencyStats struct {
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

// WindowMetrics is an aggregated view over one time window.
type WindowMetrics struct {
	EventCount  int          `json:"event_count"`
	ErrorCount  int          `json:"error_count"`
	Throughput  float64      `json:"throughput"`
	ErrorRate   float64      `json:"error_rate"`
	LatencyMS   LatencyStats `json:"latency_ms"`
	WindowLabel string       `json:"window_label"`
}

// TimeseriesPoint is a compact chart point sent to dashboard.
type TimeseriesPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Throughput  float64   `json:"throughput"`
	ErrorRate   float64   `json:"error_rate"`
	LatencyP95  float64   `json:"latency_p95"`
	ConsumerLag float64   `json:"consumer_lag"`
}

// Snapshot is the primary payload for dashboard and API clients.
type Snapshot struct {
	SchemaVersion string                   `json:"schema_version"`
	GeneratedAt   time.Time                `json:"generated_at"`
	Windows       map[string]WindowMetrics `json:"windows"`
	ConsumerLag   float64                  `json:"consumer_lag"`
	RecentPoints  []TimeseriesPoint        `json:"recent_points"`
}

// StreamEnvelope is emitted through SSE channel.
type StreamEnvelope struct {
	SchemaVersion string      `json:"schema_version"`
	EventType     string      `json:"event_type"`
	GeneratedAt   time.Time   `json:"generated_at"`
	Payload       interface{} `json:"payload"`
}
