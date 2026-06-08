package mailer_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/mailer"
)

func TestFake_Send(t *testing.T) {
	t.Parallel()

	m := mailer.NewFake()
	first := mailer.Message{From: "a@example.com", To: []string{"b@example.com"}, Subject: "first"}
	second := mailer.Message{From: "a@example.com", To: []string{"c@example.com"}, Subject: "second"}

	require.NoError(t, m.Send(context.Background(), first))
	require.NoError(t, m.Send(context.Background(), second))

	assert.Equal(t, []mailer.Message{first, second}, m.Sent())
}

func TestFake_SentReturnsCopy(t *testing.T) {
	t.Parallel()

	m := mailer.NewFake()
	require.NoError(t, m.Send(context.Background(), mailer.Message{Subject: "first"}))

	got := m.Sent()
	got[0] = mailer.Message{Subject: "mutated"}

	assert.Equal(t, []mailer.Message{{Subject: "first"}}, m.Sent())
}

func TestFake_NoSendsReturnsEmpty(t *testing.T) {
	t.Parallel()

	m := mailer.NewFake()

	assert.Empty(t, m.Sent())
}

func TestFake_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	m := mailer.NewFake()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.Send(context.Background(), mailer.Message{Subject: "evt"})
		}()
	}
	wg.Wait()

	assert.Len(t, m.Sent(), 50)
}

func TestFake_Interface(t *testing.T) {
	t.Parallel()

	var _ mailer.Mailer = mailer.NewFake()
}
