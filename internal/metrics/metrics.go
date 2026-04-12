package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once sync.Once

	ingestEventsTotal *prometheus.CounterVec
	ingestLatencyMS   prometheus.Histogram
	ingestConsumerLag prometheus.Gauge
	ingestBatchSize   prometheus.Histogram

	kafkaPublishTotal *prometheus.CounterVec
	kafkaPublishMS    prometheus.Histogram
)

func initMetrics() {
	ingestEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ingest_events_total",
		Help: "Total consumed events by ingestion status.",
	}, []string{"status"})

	ingestLatencyMS = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ingest_processing_latency_ms",
		Help:    "Event processing latency in milliseconds.",
		Buckets: []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2000, 5000},
	})

	ingestConsumerLag = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ingest_consumer_lag",
		Help: "Latest observed Kafka consumer lag in offsets.",
	})

	ingestBatchSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ingest_batch_size",
		Help:    "Batch size used by ingestion consumer.",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 200, 500},
	})

	kafkaPublishTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_publish_total",
		Help: "Total Kafka publish attempts by status.",
	}, []string{"status"})

	kafkaPublishMS = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "kafka_publish_latency_ms",
		Help:    "Kafka publish latency in milliseconds.",
		Buckets: []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000},
	})
}

func ensure() {
	once.Do(initMetrics)
}

func ObserveIngestEvent(status string, latencyMS float64) {
	ensure()
	ingestEventsTotal.WithLabelValues(status).Inc()
	ingestLatencyMS.Observe(latencyMS)
}

func ObserveIngestBatch(size int) {
	ensure()
	ingestBatchSize.Observe(float64(size))
}

func SetConsumerLag(lag int64) {
	ensure()
	if lag < 0 {
		lag = 0
	}
	ingestConsumerLag.Set(float64(lag))
}

func ObserveKafkaPublish(status string, latencyMS float64) {
	ensure()
	kafkaPublishTotal.WithLabelValues(status).Inc()
	kafkaPublishMS.Observe(latencyMS)
}

func Handler() http.Handler {
	ensure()
	return promhttp.Handler()
}
