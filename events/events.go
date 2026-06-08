package events

import "context"

// Event is implemented by all domain events.
type Event interface {
	EventType() string
}

// Handler is a function that handles a published event.
type Handler func(ctx context.Context, event Event) error

// Publisher is responsible for publishing events to subscribers.
// Implementations can be in-process (like the provided Bus) or distributed (e.g., using a message broker).
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

// Subscriber allows registration of handlers for specific event types.
type Subscriber interface {
	Subscribe(eventType string, h Handler)
}
