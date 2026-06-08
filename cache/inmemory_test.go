package cache_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/cache"
)

func TestNewInMemory(t *testing.T) {
	t.Parallel()

	t.Run("returns usable cache", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		require.NotNil(t, c)
	})

	t.Run("satisfies Cache interface", func(t *testing.T) {
		t.Parallel()

		var _ cache.Cache = cache.NewInMemory()
	})
}

func TestInMemory_GetSet(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNotFound for a missing key", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()

		_, err := c.Get(context.Background(), "missing")
		require.ErrorIs(t, err, cache.ErrNotFound)
	})

	t.Run("returns the value stored under key", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		ctx := context.Background()

		require.NoError(t, c.Set(ctx, "key", []byte("value"), 0))

		got, err := c.Get(ctx, "key")
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), got)
	})

	t.Run("set replaces an existing value", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		ctx := context.Background()

		require.NoError(t, c.Set(ctx, "key", []byte("old"), 0))
		require.NoError(t, c.Set(ctx, "key", []byte("new"), 0))

		got, err := c.Get(ctx, "key")
		require.NoError(t, err)
		assert.Equal(t, []byte("new"), got)
	})

	t.Run("stored and returned values are independent copies", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		ctx := context.Background()

		original := []byte("value")
		require.NoError(t, c.Set(ctx, "key", original, 0))
		original[0] = 'X'

		got, err := c.Get(ctx, "key")
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), got)

		got[0] = 'Y'
		got2, err := c.Get(ctx, "key")
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), got2)
	})
}

func TestInMemory_Expiry(t *testing.T) {
	t.Parallel()

	const (
		ttl  = 30 * time.Millisecond
		wait = ttl + 100*time.Millisecond
	)

	t.Run("zero ttl never expires", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		ctx := context.Background()

		require.NoError(t, c.Set(ctx, "key", []byte("value"), 0))
		time.Sleep(wait)

		_, err := c.Get(ctx, "key")
		require.NoError(t, err)
	})

	t.Run("entry is readable before it expires", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		ctx := context.Background()

		require.NoError(t, c.Set(ctx, "key", []byte("value"), time.Hour))

		_, err := c.Get(ctx, "key")
		require.NoError(t, err)
	})

	t.Run("returns ErrNotFound once the entry has expired", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		ctx := context.Background()

		require.NoError(t, c.Set(ctx, "key", []byte("value"), ttl))
		time.Sleep(wait)

		_, err := c.Get(ctx, "key")
		require.ErrorIs(t, err, cache.ErrNotFound)
	})
}

func TestInMemory_Delete(t *testing.T) {
	t.Parallel()

	t.Run("removes an existing key", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		ctx := context.Background()

		require.NoError(t, c.Set(ctx, "key", []byte("value"), 0))
		require.NoError(t, c.Delete(ctx, "key"))

		_, err := c.Get(ctx, "key")
		require.ErrorIs(t, err, cache.ErrNotFound)
	})

	t.Run("missing key returns nil", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		require.NoError(t, c.Delete(context.Background(), "missing"))
	})
}

func TestInMemory_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	t.Run("concurrent get, set and delete are race-free", func(t *testing.T) {
		t.Parallel()

		c := cache.NewInMemory()
		ctx := context.Background()

		const goroutines = 50
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := range goroutines {
			go func(i int) {
				defer wg.Done()

				key := fmt.Sprintf("key-%d", i%5)
				_ = c.Set(ctx, key, []byte("value"), time.Minute)
				_, _ = c.Get(ctx, key)
				_ = c.Delete(ctx, key)
			}(i)
		}

		wg.Wait()
	})
}
