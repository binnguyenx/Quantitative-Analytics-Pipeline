package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const SnapshotCacheKey = "telemetry:snapshot"

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(client *redis.Client, ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	return &Cache{client: client, ttl: ttl}
}

func (c *Cache) SaveSnapshot(ctx context.Context, snapshot Snapshot) error {
	if c == nil || c.client == nil {
		return nil
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	if err := c.client.Set(ctx, SnapshotCacheKey, raw, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache snapshot: %w", err)
	}
	return nil
}

func (c *Cache) GetSnapshot(ctx context.Context) (Snapshot, error) {
	if c == nil || c.client == nil {
		return Snapshot{}, redis.Nil
	}
	raw, err := c.client.Get(ctx, SnapshotCacheKey).Bytes()
	if err != nil {
		return Snapshot{}, err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("decode snapshot: %w", err)
	}
	return snapshot, nil
}
