package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/finbud/finbud-backend/internal/config"
	"github.com/finbud/finbud-backend/internal/telemetry"
	redispkg "github.com/finbud/finbud-backend/pkg/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("telemetry: failed to load config: %v", err)
	}

	rdb, err := redispkg.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("telemetry: failed to connect redis: %v", err)
	}

	svc := telemetry.NewService(telemetry.Config{
		Port:                cfg.TelemetryPort,
		MetricsPort:         cfg.TelemetryMetricsPort,
		KafkaBrokers:        cfg.KafkaBrokers,
		KafkaTopic:          cfg.KafkaTopicTelemetry,
		KafkaGroupID:        cfg.TelemetryKafkaConsumerGroupID,
		KafkaBatchSize:      cfg.TelemetryKafkaBatchSize,
		KafkaBatchWait:      time.Duration(cfg.TelemetryKafkaBatchWaitMS) * time.Millisecond,
		RedisChannel:        cfg.TelemetryRedisChannel,
		RedisSnapshotTTL:    time.Duration(cfg.TelemetrySnapshotTTLSeconds) * time.Second,
		StreamClientBufSize: cfg.TelemetryStreamBufferSize,
	}, rdb)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := svc.Run(ctx); err != nil {
		log.Fatalf("telemetry: service failed: %v", err)
	}
}
