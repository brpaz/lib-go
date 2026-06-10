package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/brpaz/lib-go/logging"
)

func newTestLogger(t *testing.T, buf *bytes.Buffer, opts ...logging.Option) *logging.Logger {
	t.Helper()
	base := []logging.Option{
		logging.WithEnvironment("test"),
		logging.WithVersion("0.0.1"),
		logging.WithRevision("abc123"),
		logging.WithFormat(logging.FormatJSON),
		logging.WithOutput(buf),
	}
	l, err := logging.NewLogger(append(base, opts...)...)
	require.NoError(t, err)
	return l
}

func decodeRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	return m
}

func TestNew_InvalidLevel(t *testing.T) {
	_, err := logging.NewLogger(logging.WithLevel("bogus"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid level")
}

func TestNew_InvalidFormat(t *testing.T) {
	_, err := logging.NewLogger(logging.WithFormat("xml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestLogger_GlobalAttrs(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, &buf)

	l.Info(context.Background(), "hello")

	rec := decodeRecord(t, &buf)
	assert.Equal(t, "test", rec["environment"])
	assert.Equal(t, "0.0.1", rec["version"])
	assert.Equal(t, "abc123", rec["revision"])
	assert.Equal(t, "hello", rec["message"])
	assert.Equal(t, "INFO", rec["level"])
	assert.NotEmpty(t, rec["timestamp"])
}

func TestLogger_RenamedKeys(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, &buf)
	l.Info(context.Background(), "key check")

	rec := decodeRecord(t, &buf)
	assert.NotContains(t, rec, "time", "slog default key 'time' should be renamed to 'timestamp'")
	assert.NotContains(t, rec, "msg", "slog default key 'msg' should be renamed to 'message'")
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, &buf, logging.WithLevel("warn"))

	l.Debug(context.Background(), "debug msg")
	l.Info(context.Background(), "info msg")
	assert.Empty(t, buf.String())

	l.Warn(context.Background(), "warn msg")
	assert.NotEmpty(t, buf.String())
}

func TestLogger_SetLevel_RuntimeChange(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, &buf, logging.WithLevel("info"))

	assert.Equal(t, slog.LevelInfo, l.GetLevel())

	l.Debug(context.Background(), "before")
	assert.Empty(t, buf.String())

	l.SetLevel(slog.LevelDebug)
	assert.Equal(t, slog.LevelDebug, l.GetLevel())

	l.Debug(context.Background(), "after")
	assert.NotEmpty(t, buf.String())
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, &buf)
	child := l.With(slog.String("user_id", "u-42"))

	child.Info(context.Background(), "enriched")

	rec := decodeRecord(t, &buf)
	assert.Equal(t, "u-42", rec["user_id"])
}

func TestLogger_With_SharedLevelVar(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, &buf, logging.WithLevel("info"))
	child := l.With(slog.String("component", "worker"))

	l.LevelVar().Set(slog.LevelDebug)
	child.Debug(context.Background(), "now visible")

	assert.NotEmpty(t, buf.String())
}

func TestLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, &buf, logging.WithFormat(logging.FormatText))

	l.Info(context.Background(), "text output")

	assert.Contains(t, buf.String(), "text output")
	// Text format should not produce a JSON object.
	assert.False(t, strings.HasPrefix(buf.String(), "{"))
}

func TestWithContext_FromContext(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(t, &buf)

	ctx := logging.WithContext(context.Background(), l)
	retrieved := logging.FromContext(ctx)
	retrieved.Info(ctx, "from context")

	rec := decodeRecord(t, &buf)
	assert.Equal(t, "from context", rec["message"])
}

func TestLogger_GlobalAttrs_Custom(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.NewLogger(
		logging.WithOutput(&buf),
		logging.WithFormat(logging.FormatJSON),
		logging.WithAttrs(
			slog.String("service", "payments"),
			slog.String("region", "eu-west-1"),
		),
	)
	require.NoError(t, err)

	l.Info(context.Background(), "charged")

	rec := decodeRecord(t, &buf)
	assert.Equal(t, "payments", rec["service"])
	assert.Equal(t, "eu-west-1", rec["region"])
}

// recordingSpan is a [trace.Span] that reports itself as recording with a
// fixed [trace.SpanContext], for testing [logging.TraceIDFromContext].
type recordingSpan struct {
	trace.Span
	sc trace.SpanContext
}

func (s recordingSpan) IsRecording() bool             { return true }
func (s recordingSpan) SpanContext() trace.SpanContext { return s.sc }

func TestTraceIDFromContext(t *testing.T) {
	t.Parallel()

	t.Run("no active span returns empty string", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, logging.TraceIDFromContext(context.Background()))
	})

	t.Run("active span returns its trace ID", func(t *testing.T) {
		t.Parallel()

		traceID, err := trace.TraceIDFromHex("0123456789abcdef0123456789abcdef")
		require.NoError(t, err)

		spanID, err := trace.SpanIDFromHex("0123456789abcdef")
		require.NoError(t, err)

		sc := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    traceID,
			SpanID:     spanID,
			TraceFlags: trace.FlagsSampled,
		})

		ctx := trace.ContextWithSpan(context.Background(), recordingSpan{sc: sc})

		assert.Equal(t, traceID.String(), logging.TraceIDFromContext(ctx))
	})
}

func TestFromContext_NoLogger_Noop(t *testing.T) {
	// Should not panic; noop logger discards everything.
	noop := logging.FromContext(context.Background())
	assert.NotNil(t, noop)
	noop.Info(context.Background(), "discarded")
}

func TestLogger_Middleware(t *testing.T) {
	type userIDKey struct{}

	userMiddleware := func(next slog.Handler) slog.Handler {
		return &userIDHandler{Handler: next, key: userIDKey{}}
	}

	var buf bytes.Buffer
	l, err := logging.NewLogger(
		logging.WithOutput(&buf),
		logging.WithFormat(logging.FormatJSON),
		logging.WithMiddleware(userMiddleware),
	)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), userIDKey{}, "u-99")
	l.Info(ctx, "action performed")

	rec := decodeRecord(t, &buf)
	assert.Equal(t, "u-99", rec["user_id"])
}

type userIDHandler struct {
	slog.Handler
	key any
}

func (h *userIDHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(h.key).(string); ok {
		r.AddAttrs(slog.String("user_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *userIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &userIDHandler{Handler: h.Handler.WithAttrs(attrs), key: h.key}
}

func (h *userIDHandler) WithGroup(name string) slog.Handler {
	return &userIDHandler{Handler: h.Handler.WithGroup(name), key: h.key}
}
