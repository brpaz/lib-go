package mailer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/mailer"
)

var _ mailer.Mailer = mailer.NewSMTP("", nil)

func TestSMTP_Send(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNoRecipients when message has no recipients", func(t *testing.T) {
		t.Parallel()

		m := mailer.NewSMTP("smtp.example.com:587", nil)
		err := m.Send(context.Background(), mailer.Message{From: "from@example.com"})
		require.ErrorIs(t, err, mailer.ErrNoRecipients)
	})

	t.Run("returns context error without dialing when context is canceled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		m := mailer.NewSMTP("smtp.example.com:587", nil)
		err := m.Send(ctx, mailer.Message{
			From: "from@example.com",
			To:   []string{"to@example.com"},
		})
		require.ErrorIs(t, err, context.Canceled)
	})
}
