package logtest_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/logging"
	"github.com/brpaz/lib-go/logging/logtest"
)

func TestNew(t *testing.T) {
	l, rec := logtest.New(t)

	l.Info(context.Background(), "hello world", slog.String("user_id", "u-1"))

	entry := rec.Last()
	assert.Equal(t, "hello world", entry["message"])
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "u-1", entry["user_id"])
	assert.NotContains(t, entry, "timestamp")
}

func TestRecorder_All(t *testing.T) {
	l, rec := logtest.New(t)

	l.Info(context.Background(), "first")
	l.Warn(context.Background(), "second")
	l.Error(context.Background(), "third")

	all := rec.All()
	require.Len(t, all, 3)
	assert.Equal(t, "first", all[0]["message"])
	assert.Equal(t, "second", all[1]["message"])
	assert.Equal(t, "third", all[2]["message"])
}

func TestRecorder_Len(t *testing.T) {
	l, rec := logtest.New(t)

	assert.Equal(t, 0, rec.Len())
	l.Info(context.Background(), "one")
	l.Info(context.Background(), "two")
	assert.Equal(t, 2, rec.Len())
}

func TestRecorder_Reset(t *testing.T) {
	l, rec := logtest.New(t)

	l.Info(context.Background(), "before")
	require.Equal(t, 1, rec.Len())

	rec.Reset()
	assert.Equal(t, 0, rec.Len())

	l.Info(context.Background(), "after")
	assert.Equal(t, 1, rec.Len())
	assert.Equal(t, "after", rec.Last()["message"])
}

func TestRecorder_LevelFiltering(t *testing.T) {
	rec := logtest.NewRecorder(t)
	l, err := logging.NewLogger(
		logging.WithOutput(rec),
		logging.WithFormat(logging.FormatJSON),
		logging.WithLevel("warn"),
	)
	require.NoError(t, err)

	l.Debug(context.Background(), "debug")
	l.Info(context.Background(), "info")
	assert.Equal(t, 0, rec.Len())

	l.Warn(context.Background(), "warn")
	assert.Equal(t, 1, rec.Len())
}

func TestRecorder_WithAttrs(t *testing.T) {
	l, rec := logtest.New(t)
	child := l.With(slog.String("component", "worker"))

	child.Info(context.Background(), "task done")

	entry := rec.Last()
	assert.Equal(t, "worker", entry["component"])
	assert.Equal(t, "task done", entry["message"])
}
