// Package redis provides a configured Redis client for caching
// simulation results and other short-lived data.
package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client, pings the server
// and returns the ready-to-use client.
func NewRedisClient(addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,              // max concurrent connections
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: ping failed: %w", err)
	}

	log.Println("redis: connected ✓")
	return client, nil
}

// DefaultCacheTTL is the standard time-to-live for cached simulation results.
const DefaultCacheTTL = 10 * time.Minute

