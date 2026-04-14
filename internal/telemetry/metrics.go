package telemetry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricsOnce sync.Once

	telemetryEventsIngested *prometheus.CounterVec
	telemetryConnectedUsers prometheus.Gauge
	telemetryDroppedFrames  prometheus.Counter
	telemetryLatencyMS      prometheus.Histogram
	telemetryConsumerLag    prometheus.Gauge
)

func initTelemetryMetrics() {
	telemetryEventsIngested = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "telemetry_events_ingested_total",
		Help: "Total telemetry events accepted by source and status.",
	}, []string{"source", "status"})

	telemetryConnectedUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_stream_clients",
		Help: "Current number of active telemetry stream clients.",
	})

	telemetryDroppedFrames = promauto.NewCounter(prometheus.CounterOpts{
		Name: "telemetry_stream_dropped_frames_total",
		Help: "Number of dropped stream frames due to slow clients.",
	})

	telemetryLatencyMS = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "telemetry_latency_ms",
		Help:    "Observed telemetry latency values from normalized events.",
		Buckets: []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000},
	})

	telemetryConsumerLag = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_consumer_lag",
		Help: "Latest telemetry consumer lag value.",
	})
}

func ensureTelemetryMetrics() {
	metricsOnce.Do(initTelemetryMetrics)
}

func observeTelemetryEvent(source, status string) {
	ensureTelemetryMetrics()
	if source == "" {
		source = "unknown"
	}
	if status == "" {
		status = "ok"
	}
	telemetryEventsIngested.WithLabelValues(source, status).Inc()
}

func observeTelemetryLatency(latencyMS float64) {
	ensureTelemetryMetrics()
	telemetryLatencyMS.Observe(latencyMS)
}

func setTelemetryConsumerLag(lag float64) {
	ensureTelemetryMetrics()
	if lag < 0 {
		lag = 0
	}
	telemetryConsumerLag.Set(lag)
}
