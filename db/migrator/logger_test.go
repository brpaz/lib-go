package migrator_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/db/migrator"
	"github.com/brpaz/lib-go/logging"
)

func newLibLogger(t *testing.T, buf *bytes.Buffer) migrator.Logger {
	t.Helper()

	l, err := logging.NewLogger(logging.WithOutput(buf))
	require.NoError(t, err)

	return migrator.NewLibLogger(l)
}

func TestNewLibLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	l := newLibLogger(t, &buf)

	l.Printf("hello %s", "world")

	assert.Contains(t, buf.String(), "hello world")
}

func TestNewNopLogger(t *testing.T) {
	t.Parallel()

	l := migrator.NewNopLogger()

	// Must not panic and must discard output.
	l.Printf("anything %d", 1)
	l.Fatalf("anything %d", 1)
}

func TestWithLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	l := newLibLogger(t, &buf)

	// Verifies WithLogger is accepted without error (no DB needed for option wiring).
	_ = migrator.WithLogger(l)
}
