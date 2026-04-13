package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	kafkago "github.com/segmentio/kafka-go"
)

type Config struct {
	Port                string
	MetricsPort         string
	KafkaBrokers        string
	KafkaTopic          string
	KafkaGroupID        string
	KafkaBatchSize      int
	KafkaBatchWait      time.Duration
	RedisChannel        string
	RedisSnapshotTTL    time.Duration
	StreamClientBufSize int
}

type Service struct {
	cfg         Config
	logger      *slog.Logger
	aggregator  *Aggregator
	hub         *Hub
	cache       *Cache
	redisClient *redis.Client
}

func NewService(cfg Config, redisClient *redis.Client) *Service {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	agg := NewAggregator(240)
	hub := NewHub(cfg.StreamClientBufSize)

	var cache *Cache
	if redisClient != nil {
		cache = NewCache(redisClient, cfg.RedisSnapshotTTL)
	}
	return &Service{
		cfg:         cfg,
		logger:      logger,
		aggregator:  agg,
		hub:         hub,
		cache:       cache,
		redisClient: redisClient,
	}
}

func (s *Service) Run(ctx context.Context) error {
	ensureTelemetryMetrics()
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	transport := NewTransport(s.aggregator, s.hub, s.cache)
	transport.RegisterRoutes(router)

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:    ":" + s.cfg.MetricsPort,
		Handler: metricsMux,
	}
	go func() {
		s.logger.Info("telemetry metrics server listening", "port", s.cfg.MetricsPort)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("metrics server failed", "error", err)
		}
	}()

	apiSrv := &http.Server{
		Addr:    ":" + s.cfg.Port,
		Handler: router,
	}
	go func() {
		s.logger.Info("telemetry api server listening", "port", s.cfg.Port)
		if err := apiSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("api server failed", "error", err)
		}
	}()

	go s.runKafkaLoop(ctx)
	go s.runRedisLoop(ctx)
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = apiSrv.Shutdown(shutdownCtx)
	_ = metricsSrv.Shutdown(shutdownCtx)
	return nil
}

func (s *Service) runKafkaLoop(ctx context.Context) {
	if s.cfg.KafkaTopic == "" {
		s.logger.Info("kafka loop disabled: no topic configured")
		return
	}
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:         splitCSV(s.cfg.KafkaBrokers),
		GroupID:         s.cfg.KafkaGroupID,
		Topic:           s.cfg.KafkaTopic,
		MinBytes:        1,
		MaxBytes:        10e6,
		CommitInterval:  0,
		ReadLagInterval: -1,
	})
	defer func() {
		_ = reader.Close()
	}()

	backoff := 200 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		batchCtx, cancel := context.WithTimeout(ctx, s.cfg.KafkaBatchWait)
		msgs := make([]kafkago.Message, 0, s.cfg.KafkaBatchSize)
		for len(msgs) < s.cfg.KafkaBatchSize {
			msg, err := reader.FetchMessage(batchCtx)
			if err != nil {
				break
			}
			msgs = append(msgs, msg)
		}
		cancel()

		if len(msgs) == 0 {
			continue
		}

		for _, msg := range msgs {
			events, err := ParseKafkaMessage(msg)
			if err != nil {
				s.logger.Warn("invalid kafka payload", "error", err)
				continue
			}
			s.pushEvents(ctx, events)
		}
		if err := reader.CommitMessages(ctx, msgs...); err != nil {
			s.logger.Warn("kafka commit failed", "error", err)
		}

		lag := reader.Lag()
		if lag >= 0 {
			s.aggregator.SetConsumerLag(float64(lag))
			s.pushEvents(ctx, []NormalizedEvent{{
				Timestamp:  time.Now().UTC(),
				Source:     "kafka",
				MetricName: "consumer_lag",
				Value:      float64(lag),
				Tags:       map[string]string{"status": "ok"},
			}})
		}
		backoff = 200 * time.Millisecond

		select {
		case <-ctx.Done():
			return
		default:
		}
		time.Sleep(backoff)
		if backoff < 2*time.Second {
			backoff *= 2
		}
	}
}

func (s *Service) runRedisLoop(ctx context.Context) {
	if s.redisClient == nil || s.cfg.RedisChannel == "" {
		return
	}
	pubsub := s.redisClient.Subscribe(ctx, s.cfg.RedisChannel)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			events, err := ParseRedisMessage(msg.Payload)
			if err != nil {
				s.logger.Warn("invalid redis telemetry payload", "error", err)
				continue
			}
			s.pushEvents(ctx, events)
		}
	}
}

func (s *Service) pushEvents(ctx context.Context, events []NormalizedEvent) {
	for _, ev := range events {
		s.aggregator.AddEvent(ev)
		telemetryEventsIngested.WithLabelValues(ev.Source).Inc()
	}
	snapshot := s.aggregator.Snapshot(time.Now().UTC())
	if err := s.cache.SaveSnapshot(ctx, snapshot); err != nil {
		s.logger.Warn("cache snapshot failed", "error", err)
	}
	env := StreamEnvelope{
		SchemaVersion: SchemaVersion,
		EventType:     "delta_update",
		GeneratedAt:   time.Now().UTC(),
		Payload:       snapshot,
	}
	payload, err := marshalSSEEnvelope(env)
	if err != nil {
		s.logger.Warn("marshal stream payload failed", "error", err)
		return
	}
	dropped := s.hub.Broadcast(payload)
	if dropped > 0 {
		telemetryDroppedFrames.Add(float64(dropped))
	}
}

func marshalSSEEnvelope(env StreamEnvelope) ([]byte, error) {
	raw, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshal envelope: %w", err)
	}
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", env.EventType, string(raw))), nil
}

func splitCSV(in string) []string {
	out := make([]string, 0, 4)
	cur := ""
	for i := 0; i < len(in); i++ {
		if in[i] == ',' {
			if cur != "" {
				out = append(out, cur)
			}
			cur = ""
			continue
		}
		if in[i] == ' ' {
			continue
		}
		cur += string(in[i])
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func EnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
