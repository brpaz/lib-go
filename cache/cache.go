package cache

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned by Get when key does not exist or has expired.
var ErrNotFound = errors.New("cache: key not found")

// Cache is a key-value store with per-entry expiration.
//
// Implementations can be in-process (like [InMemory]) or backed by an
// external store (e.g. Redis, Memcached) by providing a different
// implementation at the wiring layer. Values are stored as raw bytes —
// callers are responsible for encoding and decoding them.
type Cache interface {
	// Get returns the value stored under key, or ErrNotFound if it does not
	// exist or has expired.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores value under key, replacing any existing entry. A ttl <= 0
	// means the entry never expires.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes key. It returns nil whether or not key existed.
	Delete(ctx context.Context, key string) error
}
