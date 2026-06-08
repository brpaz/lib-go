package events_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/events"
)

// testEvent is a minimal Event implementation used across all tests.
type testEvent struct{ typ string }

func (e testEvent) EventType() string { return e.typ }

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("returns usable bus", func(t *testing.T) {
		t.Parallel()

		bus := events.New()
		require.NotNil(t, bus)
		require.NoError(t, bus.Publish(context.Background(), testEvent{"x"}))
	})

	t.Run("satisfies Publisher interface", func(t *testing.T) {
		t.Parallel()

		var _ events.Publisher = events.New()
	})

	t.Run("satisfies Subscriber interface", func(t *testing.T) {
		t.Parallel()

		var _ events.Subscriber = events.New()
	})
}

func TestBus_Publish(t *testing.T) {
	t.Parallel()

	t.Run("no handlers returns nil", func(t *testing.T) {
		t.Parallel()

		bus := events.New()
		require.NoError(t, bus.Publish(context.Background(), testEvent{"evt"}))
	})

	t.Run("calls registered handler with event", func(t *testing.T) {
		t.Parallel()

		bus := events.New()
		var received events.Event

		bus.Subscribe("evt", func(_ context.Context, e events.Event) error {
			received = e
			return nil
		})

		evt := testEvent{"evt"}
		require.NoError(t, bus.Publish(context.Background(), evt))
		assert.Equal(t, evt, received)
	})

	t.Run("passes context to handler", func(t *testing.T) {
		t.Parallel()

		type ctxKey struct{}
		bus := events.New()

		var receivedVal any
		bus.Subscribe("evt", func(ctx context.Context, _ events.Event) error {
			receivedVal = ctx.Value(ctxKey{})
			return nil
		})

		ctx := context.WithValue(context.Background(), ctxKey{}, "payload")
		require.NoError(t, bus.Publish(ctx, testEvent{"evt"}))
		assert.Equal(t, "payload", receivedVal)
	})

	t.Run("calls handlers in registration order", func(t *testing.T) {
		t.Parallel()

		bus := events.New()
		var order []int

		for i := range 3 {
			n := i
			bus.Subscribe("evt", func(_ context.Context, _ events.Event) error {
				order = append(order, n)
				return nil
			})
		}

		require.NoError(t, bus.Publish(context.Background(), testEvent{"evt"}))
		assert.Equal(t, []int{0, 1, 2}, order)
	})

	t.Run("only calls handlers for matching event type", func(t *testing.T) {
		t.Parallel()

		bus := events.New()
		var calledFoo, calledBar bool

		bus.Subscribe("foo", func(_ context.Context, _ events.Event) error {
			calledFoo = true
			return nil
		})
		bus.Subscribe("bar", func(_ context.Context, _ events.Event) error {
			calledBar = true
			return nil
		})

		require.NoError(t, bus.Publish(context.Background(), testEvent{"foo"}))

		assert.True(t, calledFoo)
		assert.False(t, calledBar)
	})

	t.Run("stops and returns error on first failing handler", func(t *testing.T) {
		t.Parallel()

		bus := events.New()
		sentinel := errors.New("handler error")
		var secondCalled bool

		bus.Subscribe("evt", func(_ context.Context, _ events.Event) error {
			return sentinel
		})
		bus.Subscribe("evt", func(_ context.Context, _ events.Event) error {
			secondCalled = true
			return nil
		})

		err := bus.Publish(context.Background(), testEvent{"evt"})

		require.ErrorIs(t, err, sentinel)
		assert.False(t, secondCalled, "second handler must not run after first returns error")
	})

	t.Run("multiple subscriptions accumulate per event type", func(t *testing.T) {
		t.Parallel()

		bus := events.New()
		var countA, countB int

		bus.Subscribe("a", func(_ context.Context, _ events.Event) error { countA++; return nil })
		bus.Subscribe("b", func(_ context.Context, _ events.Event) error { countB++; return nil })

		require.NoError(t, bus.Publish(context.Background(), testEvent{"a"}))
		require.NoError(t, bus.Publish(context.Background(), testEvent{"b"}))
		require.NoError(t, bus.Publish(context.Background(), testEvent{"a"}))

		assert.Equal(t, 2, countA)
		assert.Equal(t, 1, countB)
	})
}

func TestBus_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	t.Run("concurrent publishes are race-free", func(t *testing.T) {
		t.Parallel()

		bus := events.New()
		var mu sync.Mutex
		var count int

		bus.Subscribe("evt", func(_ context.Context, _ events.Event) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		})

		const goroutines = 50
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for range goroutines {
			go func() {
				defer wg.Done()
				_ = bus.Publish(context.Background(), testEvent{"evt"})
			}()
		}

		wg.Wait()
		assert.Equal(t, goroutines, count)
	})

	t.Run("concurrent subscribe and publish are race-free", func(t *testing.T) {
		t.Parallel()

		bus := events.New()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for range 25 {
				bus.Subscribe("evt", func(_ context.Context, _ events.Event) error { return nil })
			}
		}()

		go func() {
			defer wg.Done()
			for range 25 {
				_ = bus.Publish(context.Background(), testEvent{"evt"})
			}
		}()

		wg.Wait()
	})
}
