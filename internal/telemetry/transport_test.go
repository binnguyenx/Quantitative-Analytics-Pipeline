package telemetry

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSnapshotEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	agg := NewAggregator(5)
	agg.AddEvent(NormalizedEvent{
		Timestamp:  time.Now().UTC(),
		Source:     "test",
		MetricName: "event_processed",
		Value:      1,
	})

	hub := NewHub(4)
	transport := NewTransport(agg, hub, nil)
	r := gin.New()
	transport.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/snapshot", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var snap Snapshot
	if err := json.Unmarshal(w.Body.Bytes(), &snap); err != nil {
		t.Fatalf("failed to decode snapshot: %v", err)
	}
	if snap.SchemaVersion != SchemaVersion {
		t.Fatalf("unexpected schema version: %s", snap.SchemaVersion)
	}
}

func TestStreamEndpointSendsSnapshotEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	agg := NewAggregator(5)
	agg.AddEvent(NormalizedEvent{
		Timestamp:  time.Now().UTC(),
		Source:     "test",
		MetricName: "event_processed",
		Value:      1,
	})
	hub := NewHub(4)
	transport := NewTransport(agg, hub, nil)
	r := gin.New()
	transport.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/telemetry/stream", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("stream request failed: %v", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read stream line: %v", err)
	}
	if !strings.HasPrefix(line, "event: snapshot") {
		t.Fatalf("expected snapshot event line, got: %s", line)
	}
}
