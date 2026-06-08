package cache

import (
	"context"
	"time"
)

// Noop is a [Cache] that stores nothing. Get always returns ErrNotFound and
// Set/Delete are no-ops. Use it as a default when caching is disabled or
// irrelevant to the code under test.
type Noop struct{}

// NewNoop returns a ready-to-use Noop cache.
func NewNoop() Noop {
	return Noop{}
}

// Get always returns ErrNotFound.
func (Noop) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, ErrNotFound
}

// Set does nothing and returns nil.
func (Noop) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

// Delete does nothing and returns nil.
func (Noop) Delete(_ context.Context, _ string) error {
	return nil
}
