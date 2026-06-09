package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// contextKey is the unexported key used to store a Logger in a context.
type contextKey struct{}

// WithContext returns a copy of ctx carrying l.
func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext retrieves the Logger stored in ctx.
// If none is found it returns a no-op Logger so callers never need to nil-check.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(contextKey{}).(*Logger); ok {
		return l
	}
	return newNoop()
}

// TraceIDFromContext returns the OTEL trace ID from the active span in ctx.
// Returns an empty string when no active span is present.
func TraceIDFromContext(ctx context.Context) string {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}
