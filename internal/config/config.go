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

