package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Transport struct {
	aggregator *Aggregator
	hub        *Hub
	cache      *Cache
}

func NewTransport(aggregator *Aggregator, hub *Hub, cache *Cache) *Transport {
	ensureTelemetryMetrics()
	return &Transport{
		aggregator: aggregator,
		hub:        hub,
		cache:      cache,
	}
}

func (t *Transport) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", t.health)
	r.GET("/ready", t.readiness)
	r.GET("/telemetry/snapshot", t.snapshot)
	r.GET("/telemetry/stream", t.stream)
}

func (t *Transport) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "finbud-telemetry",
		"status":  "ok",
	})
}

func (t *Transport) readiness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "finbud-telemetry",
		"status":  "ready",
	})
}

func (t *Transport) snapshot(c *gin.Context) {
	snapshot := t.snapshotForClient(c.Request.Context())
	c.JSON(http.StatusOK, snapshot)
}

func (t *Transport) stream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	ch := t.hub.Subscribe()
	telemetryConnectedUsers.Set(float64(t.hub.Count()))
	defer func() {
		t.hub.Unsubscribe(ch)
		telemetryConnectedUsers.Set(float64(t.hub.Count()))
	}()

	initial := StreamEnvelope{
		SchemaVersion: SchemaVersion,
		EventType:     "snapshot",
		GeneratedAt:   time.Now().UTC(),
		Payload:       t.snapshotForClient(c.Request.Context()),
	}
	if err := writeSSE(c.Writer, initial.EventType, initial); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case payload := <-ch:
			if _, err := c.Writer.Write(payload); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			hb := StreamEnvelope{
				SchemaVersion: SchemaVersion,
				EventType:     "heartbeat",
				GeneratedAt:   time.Now().UTC(),
				Payload: gin.H{
					"status": "ok",
				},
			}
			if err := writeSSE(c.Writer, hb.EventType, hb); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (t *Transport) snapshotForClient(ctx context.Context) Snapshot {
	if t.cache != nil {
		if cached, err := t.cache.GetSnapshot(ctx); err == nil {
			return cached
		}
	}
	return t.aggregator.Snapshot(time.Now().UTC())
}

func writeSSE(w http.ResponseWriter, event string, payload interface{}) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(raw)); err != nil {
		return err
	}
	return nil
}
