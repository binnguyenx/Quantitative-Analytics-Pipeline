package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/finbud/finbud-backend/internal/config"
	"github.com/finbud/finbud-backend/internal/metrics"
	"github.com/finbud/finbud-backend/internal/repository"
	"github.com/finbud/finbud-backend/internal/service"
	"github.com/finbud/finbud-backend/pkg/database"
	kafkapkg "github.com/finbud/finbud-backend/pkg/kafka"
	redispkg "github.com/finbud/finbud-backend/pkg/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("consumer: failed to load config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("consumer: failed to connect postgres: %v", err)
	}

	rdb, err := redispkg.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("consumer: failed to connect redis: %v", err)
	}

	consumer := kafkapkg.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopicProfileUpdated, cfg.KafkaConsumerGroupID)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("consumer: close error: %v", err)
		}
	}()

	dlqProducer := kafkapkg.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopicProfileDLQ)
	defer func() {
		if err := dlqProducer.Close(); err != nil {
			log.Printf("consumer: dlq producer close error: %v", err)
		}
	}()

	ingestRepo := repository.NewIngestedEventRepository(db)
	ingestService := service.NewIngestionService(
		ingestRepo,
		rdb,
		dlqProducer,
		cfg.KafkaMaxRetries,
		time.Duration(cfg.KafkaRetryBackoffMS)*time.Millisecond,
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	metricsSrv := &http.Server{
		Addr:    ":" + cfg.ConsumerMetricsPort,
		Handler: mux,
	}
	go func() {
		log.Printf("consumer: metrics listening on :%s/metrics", cfg.ConsumerMetricsPort)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("consumer: metrics server failed: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	log.Printf("consumer: consuming topic=%s group=%s", cfg.KafkaTopicProfileUpdated, cfg.KafkaConsumerGroupID)

	for {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = metricsSrv.Shutdown(shutdownCtx)
			log.Println("consumer: shutdown complete")
			return
		default:
		}

		msgs, err := consumer.FetchBatch(ctx, cfg.KafkaBatchSize, time.Duration(cfg.KafkaBatchWaitMS)*time.Millisecond)
		if err != nil {
			log.Printf("consumer: fetch error: %v", err)
			time.Sleep(300 * time.Millisecond)
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		if err := ingestService.ProcessBatch(ctx, msgs); err != nil {
			log.Printf("consumer: process batch error: %v", err)
			time.Sleep(300 * time.Millisecond)
			continue
		}

		if err := consumer.CommitBatch(ctx, msgs); err != nil {
			log.Printf("consumer: commit error: %v", err)
			time.Sleep(300 * time.Millisecond)
			continue
		}
	}
}
