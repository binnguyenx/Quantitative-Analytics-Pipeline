// Package config centralises all application configuration.
// Every value is sourced from environment variables so the app
// stays 12-factor compliant and easy to deploy in containers.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds every tuneable knob for FinBud.
type Config struct {
	// ── Server ──────────────────────────────────────
	ServerPort string
	GinMode    string

	// ── PostgreSQL ──────────────────────────────────
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string

	// ── Redis ───────────────────────────────────────
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// ── Kafka ───────────────────────────────────────
	KafkaBrokers             string
	KafkaTopicProfileUpdated string
	KafkaTopicProfileDLQ     string
	KafkaConsumerGroupID     string
	KafkaBatchSize           int
	KafkaBatchWaitMS         int
	KafkaMaxRetries          int
	KafkaRetryBackoffMS      int

	// ── Metrics / Ingestion ─────────────────────────
	ConsumerMetricsPort string
}

// Load reads a .env file (if present) and then populates Config
// from environment variables.  Missing vars fall back to sane defaults.
func Load() (*Config, error) {
	// Best-effort: .env is optional (CI / Docker may inject env directly).
	_ = godotenv.Load()

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid REDIS_DB value: %w", err)
	}
	kafkaBatchSize, err := strconv.Atoi(getEnv("KAFKA_CONSUMER_BATCH_SIZE", "100"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid KAFKA_CONSUMER_BATCH_SIZE value: %w", err)
	}
	kafkaBatchWaitMS, err := strconv.Atoi(getEnv("KAFKA_CONSUMER_BATCH_WAIT_MS", "500"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid KAFKA_CONSUMER_BATCH_WAIT_MS value: %w", err)
	}
	kafkaMaxRetries, err := strconv.Atoi(getEnv("KAFKA_CONSUMER_MAX_RETRIES", "3"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid KAFKA_CONSUMER_MAX_RETRIES value: %w", err)
	}
	kafkaRetryBackoffMS, err := strconv.Atoi(getEnv("KAFKA_CONSUMER_RETRY_BACKOFF_MS", "200"))
	if err != nil {
		return nil, fmt.Errorf("config: invalid KAFKA_CONSUMER_RETRY_BACKOFF_MS value: %w", err)
	}

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		GinMode:    getEnv("GIN_MODE", "debug"),

		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "finbud"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "finbud_secret"),
		PostgresDB:       getEnv("POSTGRES_DB", "finbud_db"),
		PostgresSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		KafkaBrokers:             getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopicProfileUpdated: getEnv("KAFKA_TOPIC_PROFILE_UPDATED", "EVENT_PROFILE_UPDATED"),
		KafkaTopicProfileDLQ:     getEnv("KAFKA_TOPIC_PROFILE_DLQ", "EVENT_PROFILE_UPDATED_DLQ"),
		KafkaConsumerGroupID:     getEnv("KAFKA_CONSUMER_GROUP_ID", "finbud-ingestion-consumer"),
		KafkaBatchSize:           kafkaBatchSize,
		KafkaBatchWaitMS:         kafkaBatchWaitMS,
		KafkaMaxRetries:          kafkaMaxRetries,
		KafkaRetryBackoffMS:      kafkaRetryBackoffMS,

		ConsumerMetricsPort: getEnv("CONSUMER_METRICS_PORT", "2113"),
	}, nil
}

// PostgresDSN builds a GORM-compatible PostgreSQL connection string.
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresHost, c.PostgresPort, c.PostgresUser,
		c.PostgresPassword, c.PostgresDB, c.PostgresSSLMode,
	)
}

// getEnv returns the environment variable value or a fallback.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

