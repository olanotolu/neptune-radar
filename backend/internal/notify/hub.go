package notify

import (
	"sync"
	"time"
)

// Event is one pipeline event pushed to SSE subscribers.
type Event struct {
	Type string      // "couple_detected", "stage_transition", "action_created", "dlq_item", "alert"
	Data interface{} // JSON-serializable payload
	Time time.Time
}

// Hub is a simple in-memory pub/sub hub for SSE live updates.
// Publish is non-blocking: a slow subscriber's buffer fills and subsequent
// events are dropped for that subscriber rather than stalling the pipeline.
// ponytail: ceiling — in-memory, per-process. No persistence, no fan-out
// across replicas. Fine for a single-operator dashboard; upgrade to Redis
// pub/sub if multiple replicas need to share events.
type Hub struct {
	subscribers map[chan Event]struct{}
	mu          sync.RWMutex
}

const subscriberBufferSize = 64

func NewHub() *Hub {
	return &Hub{subscribers: make(map[chan Event]struct{})}
}

// Subscribe returns a buffered channel. The caller MUST Unsubscribe when done
// to avoid leaking the channel (and blocking the hub's map growth).
func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, subscriberBufferSize)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	delete(h.subscribers, ch)
	h.mu.Unlock()
	close(ch)
}

// Publish delivers the event to every subscriber. Non-blocking: if a
// subscriber's buffer is full the event is dropped for that subscriber only.
func (h *Hub) Publish(e Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subscribers {
		select {
		case ch <- e:
		default:
			// Buffer full — drop for this subscriber. A slow client must
			// never block the pipeline.
		}
	}
}
