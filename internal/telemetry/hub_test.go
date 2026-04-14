package telemetry

import "testing"

func TestHubBroadcastBackpressureDropCount(t *testing.T) {
	hub := NewHub(1)
	ch := hub.Subscribe()
	defer hub.Unsubscribe(ch)

	// Fill subscriber buffer to trigger drop path.
	ch <- []byte("first")
	dropped := hub.Broadcast([]byte("second"))
	if dropped < 0 {
		t.Fatalf("expected non-negative dropped count, got %d", dropped)
	}
}
