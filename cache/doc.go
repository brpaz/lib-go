// Package cache provides a key-value [Cache] abstraction with per-entry
// expiration, plus implementations for common scenarios.
//
// # Storing and retrieving values
//
// [Cache] stores raw bytes under a string key. Callers are responsible for
// encoding and decoding their own values:
//
//	data, _ := json.Marshal(profile)
//	_ = c.Set(ctx, "profile:"+userID, data, time.Hour)
//
//	raw, err := c.Get(ctx, "profile:"+userID)
//	if errors.Is(err, cache.ErrNotFound) {
//	    // load from the source of truth and repopulate the cache
//	}
//
// # Expiration
//
// Set takes a ttl. A ttl <= 0 means the entry never expires:
//
//	_ = c.Set(ctx, "session:"+token, data, 30*time.Minute) // expires
//	_ = c.Set(ctx, "config:flags", data, 0)                // never expires
//
// # Testing
//
// Pass [Noop] when a component needs a [Cache] but the test doesn't care
// about caching — every Get misses and every Set/Delete is discarded:
//
//	svc := NewProfileService(cache.Noop{})
//
// Use [InMemory] when a test needs to assert on hits, misses or values
// without a real backend:
//
//	c := cache.NewInMemory()
//
//	_ = c.Set(ctx, "key", []byte("value"), time.Minute)
//	got, err := c.Get(ctx, "key")
//	require.NoError(t, err)
//	assert.Equal(t, []byte("value"), got)
//
// # Swapping the implementation
//
// Code should depend on the [Cache] interface rather than a concrete type.
// This allows swapping [InMemory] for a distributed backend (e.g. Redis,
// Memcached) at the wiring layer without touching the rest of the codebase.
package cache
