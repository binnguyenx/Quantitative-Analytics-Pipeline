package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/finbud/finbud-backend/internal/config"
	redispkg "github.com/finbud/finbud-backend/pkg/redis"
)

type message struct {
	Event      string                 `json:"event"`
	Timestamp  string                 `json:"timestamp"`
	MetricName string                 `json:"metric_name"`
	Value      float64                `json:"value"`
	Source     string                 `json:"source"`
	Tags       map[string]interface{} `json:"tags"`
	Data       map[string]interface{} `json:"data"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("telemetry-publisher: load config failed: %v", err)
	}
	rdb, err := redispkg.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("telemetry-publisher: connect redis failed: %v", err)
	}

	ctx := context.Background()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	log.Printf("telemetry-publisher: publishing sample metrics to redis channel=%s", cfg.TelemetryRedisChannel)

	lag := 5.0
	for range ticker.C {
		lag += rand.Float64()*2 - 1
		if lag < 0 {
			lag = 0
		}

		status := "ok"
		if rand.Float64() < 0.12 {
			status = "error"
		}
		latency := 50 + rand.Float64()*250

		m := message{
			Event:      "TELEMETRY_METRIC",
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			MetricName: "latency_ms",
			Value:      latency,
			Source:     "telemetry-publisher",
			Tags: map[string]interface{}{
				"status": status,
			},
			Data: map[string]interface{}{
				"consumer_lag": lag,
				"status":       status,
			},
		}

		raw, _ := json.Marshal(m)
		if err := rdb.Publish(ctx, cfg.TelemetryRedisChannel, raw).Err(); err != nil {
			log.Printf("telemetry-publisher: publish error: %v", err)
		}
	}
}
