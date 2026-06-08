package mailer

import "context"

// Noop is a [Mailer] that discards every message. Use it as a default when
// email sending is disabled or irrelevant to the code under test.
type Noop struct{}

// NewNoop returns a ready-to-use Noop mailer.
func NewNoop() Noop {
	return Noop{}
}

// Send does nothing and returns nil.
func (Noop) Send(_ context.Context, _ Message) error {
	return nil
}
