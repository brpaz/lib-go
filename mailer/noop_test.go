package mailer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/mailer"
)

func TestNewNoop(t *testing.T) {
	t.Parallel()

	require.Equal(t, mailer.Noop{}, mailer.NewNoop())
}

func TestNoop_Send(t *testing.T) {
	t.Parallel()

	err := mailer.Noop{}.Send(context.Background(), mailer.Message{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "subject",
		Body:    "body",
	})
	require.NoError(t, err)
}

func TestNoop_Interface(t *testing.T) {
	t.Parallel()

	var _ mailer.Mailer = mailer.Noop{}
}
