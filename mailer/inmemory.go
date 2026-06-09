package mailer

import (
	"context"
	"sync"
)

// InMemory records every sent message instead of delivering it. Use it in tests
// to assert which emails a service sent, without wiring a real transport.
// Safe for concurrent use.
type InMemory struct {
	mu   sync.Mutex
	sent []Message
}

// NewInMemory creates an empty InMemory mailer.
func NewInMemory() *InMemory {
	return &InMemory{}
}

// Send records msg and returns nil.
func (m *InMemory) Send(_ context.Context, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sent = append(m.sent, msg)

	return nil
}

// Sent returns a copy of every message recorded so far, in send order.
func (m *InMemory) Sent() []Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]Message, len(m.sent))
	copy(out, m.sent)

	return out
}
