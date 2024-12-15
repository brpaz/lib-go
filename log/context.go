package log

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type ctxKeyType struct{}

var loggerCtxKey = ctxKeyType{}

// FromContext retrieves the logger instance from the provided context.
// If no logger is present in the context, it falls back to the global logger.
func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return L()
	}

	logger, ok := ctx.Value(loggerCtxKey).(Logger)
	if !ok || logger == nil {
		return L()
	}

	return logger
}

// ContextWithLogger returns a new context with the provided logger instance.
// This is useful for associating a specific logger with the given context.
func ContextWithLogger(parentCtx context.Context, logger Logger) context.Context {
	if parentCtx == nil {
		parentCtx = context.TODO()
	}
	return context.WithValue(parentCtx, loggerCtxKey, logger)
}

// ExtractTraceIDFieldsFromContext extracts trace-related fields from the given context
// if OpenTelemetry tracing information is available.
// The extracted fields include "traceId" and "spanId".
// Returns an empty slice if no valid trace context is found.
func ExtractTraceIDFieldsFromContext(ctx context.Context) []Field {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	if !spanContext.IsValid() {
		return []Field{}
	}

	return []Field{
		String("traceId", spanContext.TraceID().String()),
		String("spanId", spanContext.SpanID().String()),
	}
}
