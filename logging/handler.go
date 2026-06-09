package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// OtelMiddleware returns a [Middleware] that injects trace_id and span_id from
// the active OpenTelemetry span into every log record.
// Add it via [WithMiddleware] when your application uses OpenTelemetry tracing.
func OtelMiddleware() Middleware {
	return func(next slog.Handler) slog.Handler {
		return &otelHandler{Handler: next}
	}
}

type otelHandler struct {
	slog.Handler
}

func (h *otelHandler) Handle(ctx context.Context, r slog.Record) error {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		sc := span.SpanContext()
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *otelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &otelHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *otelHandler) WithGroup(name string) slog.Handler {
	return &otelHandler{Handler: h.Handler.WithGroup(name)}
}
