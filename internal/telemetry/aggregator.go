package telemetry

import (
	"math"
	"sort"
	"sync"
	"time"
)

type Aggregator struct {
	mu           sync.RWMutex
	events       []NormalizedEvent
	recentPoints []TimeseriesPoint
	consumerLag  float64
	maxPoints    int
}

func NewAggregator(maxPoints int) *Aggregator {
	if maxPoints <= 0 {
		maxPoints = 180
	}
	return &Aggregator{
		events:       make([]NormalizedEvent, 0, 512),
		recentPoints: make([]TimeseriesPoint, 0, maxPoints),
		maxPoints:    maxPoints,
	}
}

func (a *Aggregator) AddEvent(event NormalizedEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	ts := event.Timestamp.UTC()
	if ts.IsZero() {
		ts = time.Now().UTC()
		event.Timestamp = ts
	}

	a.events = append(a.events, event)
	a.pruneEventsLocked(ts)
	snapshot := a.snapshotLocked(ts)
	w1 := snapshot.Windows["1m"]
	a.recentPoints = append(a.recentPoints, TimeseriesPoint{
		Timestamp:   ts,
		Throughput:  w1.Throughput,
		ErrorRate:   w1.ErrorRate,
		LatencyP95:  w1.LatencyMS.P95,
		ConsumerLag: a.consumerLag,
	})
	if len(a.recentPoints) > a.maxPoints {
		a.recentPoints = a.recentPoints[len(a.recentPoints)-a.maxPoints:]
	}
}

func (a *Aggregator) SetConsumerLag(lag float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if lag < 0 {
		lag = 0
	}
	a.consumerLag = lag
}

func (a *Aggregator) Snapshot(now time.Time) Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	a.pruneEventsLocked(now)
	return a.snapshotLocked(now)
}

func (a *Aggregator) pruneEventsLocked(now time.Time) {
	cutoff := now.Add(-15 * time.Minute)
	idx := 0
	for idx < len(a.events) && a.events[idx].Timestamp.Before(cutoff) {
		idx++
	}
	if idx > 0 {
		a.events = append([]NormalizedEvent{}, a.events[idx:]...)
	}
}

func (a *Aggregator) snapshotLocked(now time.Time) Snapshot {
	windows := map[string]WindowMetrics{
		"1m":  a.computeWindowLocked(now, 1*time.Minute, "1m"),
		"5m":  a.computeWindowLocked(now, 5*time.Minute, "5m"),
		"15m": a.computeWindowLocked(now, 15*time.Minute, "15m"),
	}
	recent := append([]TimeseriesPoint{}, a.recentPoints...)
	return Snapshot{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   now.UTC(),
		Windows:       windows,
		ConsumerLag:   a.consumerLag,
		RecentPoints:  recent,
	}
}

func (a *Aggregator) computeWindowLocked(now time.Time, window time.Duration, label string) WindowMetrics {
	cutoff := now.Add(-window)
	latencies := make([]float64, 0, 64)
	total := 0
	errors := 0

	for _, ev := range a.events {
		if ev.Timestamp.Before(cutoff) {
			continue
		}
		total++
		if status, ok := ev.Tags["status"]; ok && status == "error" {
			errors++
		}
		if ev.MetricName == "error_count" {
			errors += int(ev.Value)
		}
		if ev.MetricName == "latency_ms" {
			latencies = append(latencies, ev.Value)
		}
		if ev.MetricName == "consumer_lag" {
			a.consumerLag = ev.Value
		}
	}

	throughput := float64(total) / window.Seconds()
	errorRate := 0.0
	if total > 0 {
		errorRate = float64(errors) / float64(total) * 100
	}

	return WindowMetrics{
		EventCount:  total,
		ErrorCount:  errors,
		Throughput:  round(throughput, 4),
		ErrorRate:   round(errorRate, 4),
		LatencyMS:   percentileStats(latencies),
		WindowLabel: label,
	}
}

func percentileStats(values []float64) LatencyStats {
	if len(values) == 0 {
		return LatencyStats{}
	}
	cp := append([]float64{}, values...)
	sort.Float64s(cp)
	return LatencyStats{
		P50: round(percentile(cp, 0.50), 4),
		P95: round(percentile(cp, 0.95), 4),
		P99: round(percentile(cp, 0.99), 4),
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	pos := p * float64(len(sorted)-1)
	l := int(math.Floor(pos))
	u := int(math.Ceil(pos))
	if l == u {
		return sorted[l]
	}
	w := pos - float64(l)
	return sorted[l]*(1-w) + sorted[u]*w
}

func round(v float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(v*p) / p
}
