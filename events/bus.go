package events

import (
	"context"
	"sync"
)

// Bus is a synchronous, in-process event bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// New returns a ready-to-use Bus.
func New() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

// Subscribe registers h to be called whenever an event of the given type is published.
func (b *Bus) Subscribe(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

// Publish calls all handlers registered for event.EventType() in order.
// Returns the first handler error encountered; remaining handlers are not called.
func (b *Bus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	hs := b.handlers[event.EventType()]
	b.mu.RUnlock()

	for _, h := range hs {
		if err := h(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
