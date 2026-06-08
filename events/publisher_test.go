package events_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/events"
)

func TestNewNoopPublisher(t *testing.T) {
	t.Parallel()

	require.Equal(t, events.NoopPublisher{}, events.NewNoopPublisher())
}

func TestNoopPublisher_Publish(t *testing.T) {
	t.Parallel()

	require.NoError(t, events.NoopPublisher{}.Publish(context.Background(), testEvent{"evt"}))
}

func TestNoopPublisher_Interface(t *testing.T) {
	t.Parallel()

	var _ events.Publisher = events.NoopPublisher{}
}

func TestFakePublisher_Publish(t *testing.T) {
	t.Parallel()

	p := events.NewFakePublisher()
	first := testEvent{"first"}
	second := testEvent{"second"}

	require.NoError(t, p.Publish(context.Background(), first))
	require.NoError(t, p.Publish(context.Background(), second))

	assert.Equal(t, []events.Event{first, second}, p.Published())
}

func TestFakePublisher_PublishedReturnsCopy(t *testing.T) {
	t.Parallel()

	p := events.NewFakePublisher()
	require.NoError(t, p.Publish(context.Background(), testEvent{"first"}))

	got := p.Published()
	got[0] = testEvent{"mutated"}

	assert.Equal(t, []events.Event{testEvent{"first"}}, p.Published())
}

func TestFakePublisher_NoPublishesReturnsEmpty(t *testing.T) {
	t.Parallel()

	p := events.NewFakePublisher()

	assert.Empty(t, p.Published())
}

func TestFakePublisher_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	p := events.NewFakePublisher()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.Publish(context.Background(), testEvent{"evt"})
		}()
	}
	wg.Wait()

	assert.Len(t, p.Published(), 50)
}

func TestFakePublisher_Interface(t *testing.T) {
	t.Parallel()

	var _ events.Publisher = events.NewFakePublisher()
}
