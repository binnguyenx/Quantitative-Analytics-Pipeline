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
)

func initTelemetryMetrics() {
	telemetryEventsIngested = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "telemetry_events_ingested_total",
		Help: "Total telemetry events accepted by source.",
	}, []string{"source"})

	telemetryConnectedUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_stream_clients",
		Help: "Current number of active telemetry stream clients.",
	})

	telemetryDroppedFrames = promauto.NewCounter(prometheus.CounterOpts{
		Name: "telemetry_stream_dropped_frames_total",
		Help: "Number of dropped stream frames due to slow clients.",
	})
}

func ensureTelemetryMetrics() {
	metricsOnce.Do(initTelemetryMetrics)
}
