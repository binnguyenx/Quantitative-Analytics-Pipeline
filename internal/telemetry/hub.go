package telemetry

import "sync"

type Hub struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
	bufferSize  int
}

func NewHub(bufferSize int) *Hub {
	if bufferSize <= 0 {
		bufferSize = 32
	}
	return &Hub{
		subscribers: make(map[chan []byte]struct{}),
		bufferSize:  bufferSize,
	}
}

func (h *Hub) Subscribe() chan []byte {
	ch := make(chan []byte, h.bufferSize)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	if _, ok := h.subscribers[ch]; ok {
		delete(h.subscribers, ch)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(payload []byte) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	dropped := 0
	for ch := range h.subscribers {
		select {
		case ch <- payload:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- payload:
			default:
				dropped++
			}
		}
	}
	return dropped
}

func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers)
}
