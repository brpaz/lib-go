package events

import (
	"context"
	"sync"
)

// NoopPublisher discards every event. Use it as a default [Publisher] when
// events are disabled or irrelevant to the code under test.
type NoopPublisher struct{}

// NewNoopPublisher returns a ready-to-use NoopPublisher.
func NewNoopPublisher() NoopPublisher {
	return NoopPublisher{}
}

// Publish does nothing and returns nil.
func (NoopPublisher) Publish(_ context.Context, _ Event) error {
	return nil
}

// FakePublisher records every published event instead of dispatching it. Use
// it in tests to assert which events a service published, without wiring a
// [Bus] and handlers. Safe for concurrent use.
type FakePublisher struct {
	mu        sync.Mutex
	published []Event
}

// NewFakePublisher creates an empty FakePublisher.
func NewFakePublisher() *FakePublisher {
	return &FakePublisher{}
}

// Publish records event and returns nil.
func (p *FakePublisher) Publish(_ context.Context, event Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.published = append(p.published, event)

	return nil
}

// Published returns a copy of every event recorded so far, in publish order.
func (p *FakePublisher) Published() []Event {
	p.mu.Lock()
	defer p.mu.Unlock()

	out := make([]Event, len(p.published))
	copy(out, p.published)

	return out
}
