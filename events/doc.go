// Package events provides a synchronous, in-process event bus for decoupled
// communication between components.
//
// The [Bus] dispatches events to registered handlers in the order they were
// subscribed. Publishing is synchronous — Publish returns only after every
// handler has run (or the first one returns an error).
//
// # Defining events
//
// Each event is a small, immutable type that implements [Event] by returning
// a stable string key:
//
//	type OrderPlacedEvent struct {
//	    OrderID string
//	    Total   int64
//	}
//
//	func (e OrderPlacedEvent) EventType() string { return "order.placed" }
//
// # Publishing
//
// Services hold a [Publisher] (or the concrete *[Bus]) and call Publish after
// the primary operation succeeds:
//
//	_ = s.publisher.Publish(ctx, OrderPlacedEvent{
//	    OrderID: order.ID,
//	    Total:   order.Total,
//	})
//
// Use _ = to swallow the error when handler failures must not abort the primary
// operation. Return the error when cleanup is a hard requirement.
//
// # Subscribing
//
// Register handlers against the shared [Bus] during startup:
//
//	bus.Subscribe("order.placed", func(ctx context.Context, e events.Event) error {
//	    evt := e.(OrderPlacedEvent)
//	    return invoicing.CreateDraft(ctx, evt.OrderID, evt.Total)
//	})
//
// # Testing
//
// Pass [NoopPublisher] when a component needs a [Publisher] but the test
// doesn't care about events, and [FakePublisher] to assert which events were
// published:
//
//	pub := events.NewFakePublisher()
//	svc := NewOrderService(pub)
//
//	_ = svc.Place(ctx, order)
//	assert.Equal(t, []events.Event{
//	    OrderPlacedEvent{OrderID: order.ID, Total: order.Total},
//	}, pub.Published())
//
// # Swapping the implementation
//
// Code should depend on the [Publisher] interface rather than *[Bus] directly.
// This allows swapping to a distributed broker (e.g. NATS, Kafka) by providing
// a different implementation at the wiring layer without touching the rest of
// the codebase.
package events
