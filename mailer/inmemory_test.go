package mailer_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/mailer"
)

var _ mailer.Mailer = mailer.NewInMemory()

func TestInMemory_Send(t *testing.T) {
	t.Parallel()

	t.Run("records sent messages in order", func(t *testing.T) {
		t.Parallel()

		m := mailer.NewInMemory()
		first := mailer.Message{From: "a@example.com", To: []string{"b@example.com"}, Subject: "first"}
		second := mailer.Message{From: "a@example.com", To: []string{"c@example.com"}, Subject: "second"}

		require.NoError(t, m.Send(context.Background(), first))
		require.NoError(t, m.Send(context.Background(), second))

		assert.Equal(t, []mailer.Message{first, second}, m.Sent())
	})

	t.Run("sent returns empty before any sends", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, mailer.NewInMemory().Sent())
	})

	t.Run("sent returns a copy", func(t *testing.T) {
		t.Parallel()

		m := mailer.NewInMemory()
		require.NoError(t, m.Send(context.Background(), mailer.Message{Subject: "first"}))

		got := m.Sent()
		got[0] = mailer.Message{Subject: "mutated"}

		assert.Equal(t, []mailer.Message{{Subject: "first"}}, m.Sent())
	})

	t.Run("safe for concurrent use", func(t *testing.T) {
		t.Parallel()

		m := mailer.NewInMemory()
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
	})
}
