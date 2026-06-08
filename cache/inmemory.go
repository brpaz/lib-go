package cache

import (
	"context"
	"sync"
	"time"
)

// InMemory is an in-process [Cache] backed by a map. Entries expire lazily —
// an expired entry is treated as missing and removed the next time it is
// read, and is replaced outright on the next write to the same key. Safe for
// concurrent use.
type InMemory struct {
	mu      sync.RWMutex
	entries map[string]inMemoryEntry
}

type inMemoryEntry struct {
	value     []byte
	expiresAt time.Time // zero means no expiry
}

func (e inMemoryEntry) expired(now time.Time) bool {
	return !e.expiresAt.IsZero() && now.After(e.expiresAt)
}

// NewInMemory returns a ready-to-use InMemory cache.
func NewInMemory() *InMemory {
	return &InMemory{
		entries: make(map[string]inMemoryEntry),
	}
}

// Get implements [Cache].
func (c *InMemory) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		return nil, ErrNotFound
	}

	if entry.expired(time.Now()) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()

		return nil, ErrNotFound
	}

	out := make([]byte, len(entry.value))
	copy(out, entry.value)

	return out, nil
}

// Set implements [Cache].
func (c *InMemory) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	stored := make([]byte, len(value))
	copy(stored, value)

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.mu.Lock()
	c.entries[key] = inMemoryEntry{value: stored, expiresAt: expiresAt}
	c.mu.Unlock()

	return nil
}

// Delete implements [Cache].
func (c *InMemory) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()

	return nil
}
