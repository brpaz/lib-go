package mailer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/mailer"
)

func TestNewSMTP(t *testing.T) {
	t.Parallel()

	t.Run("returns usable mailer", func(t *testing.T) {
		t.Parallel()

		m := mailer.NewSMTP("smtp.example.com:587", nil)
		require.NotNil(t, m)
	})

	t.Run("satisfies Mailer interface", func(t *testing.T) {
		t.Parallel()

		var _ mailer.Mailer = mailer.NewSMTP("smtp.example.com:587", nil)
	})
}

func TestSMTP_Send(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNoRecipients when the message has no recipients", func(t *testing.T) {
		t.Parallel()

		m := mailer.NewSMTP("smtp.example.com:587", nil)

		err := m.Send(context.Background(), mailer.Message{From: "from@example.com"})
		require.ErrorIs(t, err, mailer.ErrNoRecipients)
	})

	t.Run("returns the context error without dialing when context is canceled", func(t *testing.T) {
		t.Parallel()

		m := mailer.NewSMTP("smtp.example.com:587", nil)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := m.Send(ctx, mailer.Message{
			From: "from@example.com",
			To:   []string{"to@example.com"},
		})
		require.ErrorIs(t, err, context.Canceled)
	})
}
