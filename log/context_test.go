package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"

	"github.com/brpaz/lib-go/log"
)

func TestFromContext(t *testing.T) {
	t.Parallel()

	t.Run("nil context should return global logger", func(t *testing.T) {
		t.Parallel()
		globalLogger := log.NewInMemory(log.LevelDebug)
		log.ReplaceGlobals(globalLogger)()
		defer log.ReplaceGlobals(globalLogger)()

		//nolint:staticcheck
		logger := log.FromContext(nil)
		assert.Equal(t, globalLogger, logger)
	})

	t.Run("context with no logger should return global logger", func(t *testing.T) {
		t.Parallel()
		globalLogger := log.NewInMemory(log.LevelDebug)
		log.ReplaceGlobals(globalLogger)()
		defer log.ReplaceGlobals(globalLogger)()

		ctx := context.Background()
		logger := log.FromContext(ctx)
		assert.Equal(t, globalLogger, logger)
	})

	t.Run("context with logger should return the logger", func(t *testing.T) {
		t.Parallel()
		logger := log.NewInMemory(log.LevelDebug)
		ctx := log.ContextWithLogger(context.Background(), logger)

		retrievedLogger := log.FromContext(ctx)
		assert.Equal(t, logger, retrievedLogger)
	})
}

func TestContextWithLogger(t *testing.T) {
	t.Parallel()

	t.Run("returns a context containing the provided logger", func(t *testing.T) {
		t.Parallel()

		logger := log.NewInMemory(log.LevelDebug)
		ctx := log.ContextWithLogger(context.Background(), logger)

		retrievedLogger := log.FromContext(ctx)
		assert.Equal(t, logger, retrievedLogger)
	})
}

func TestExtractTraceFieldsFromContext(t *testing.T) {
	t.Parallel()

	t.Run("no valid trace context should return an empty slice", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		fields := log.ExtractTraceIDFieldsFromContext(ctx)
		assert.Empty(t, fields)
	})

	t.Run("valid trace context should return traceId and spanId fields", func(t *testing.T) {
		t.Parallel()
		mockTraceID := trace.TraceID([16]byte{1, 2, 3, 4})
		mockSpanID := trace.SpanID([8]byte{5, 6, 7, 8})

		spanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: mockTraceID,
			SpanID:  mockSpanID,
		})
		ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

		fields := log.ExtractTraceIDFieldsFromContext(ctx)

		assert.Len(t, fields, 2)
		assert.Contains(t, fields, log.String("traceId", mockTraceID.String()))
		assert.Contains(t, fields, log.String("spanId", mockSpanID.String()))
	})
}
