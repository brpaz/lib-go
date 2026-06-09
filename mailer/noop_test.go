package mailer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/mailer"
)

var _ mailer.Mailer = mailer.Noop{}

func TestNoop_Send(t *testing.T) {
	t.Parallel()

	t.Run("discards message and returns nil", func(t *testing.T) {
		t.Parallel()

		err := mailer.NewNoop().Send(context.Background(), mailer.Message{
			From:    "from@example.com",
			To:      []string{"to@example.com"},
			Subject: "subject",
			Body:    "body",
		})
		require.NoError(t, err)
	})
}
