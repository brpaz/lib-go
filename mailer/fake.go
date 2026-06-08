package mailer

import (
	"context"
	"sync"
)

// Fake records every sent message instead of delivering it. Use it in tests
// to assert which emails a service sent, without wiring a real transport.
// Safe for concurrent use.
type Fake struct {
	mu   sync.Mutex
	sent []Message
}

// NewFake creates an empty Fake mailer.
func NewFake() *Fake {
	return &Fake{}
}

// Send records msg and returns nil.
func (m *Fake) Send(_ context.Context, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sent = append(m.sent, msg)

	return nil
}

// Sent returns a copy of every message recorded so far, in send order.
func (m *Fake) Sent() []Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]Message, len(m.sent))
	copy(out, m.sent)

	return out
}
