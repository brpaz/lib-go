package tracingtest

import (
	"context"
	"crypto/rand"

	"go.opentelemetry.io/otel/trace"
)

// fakeSpan is a minimal trace.Span that returns a known SpanContext.
type fakeSpan struct {
	trace.Span
	sc trace.SpanContext
}

func (s fakeSpan) SpanContext() trace.SpanContext { return s.sc }
func (s fakeSpan) IsRecording() bool              { return s.sc.IsValid() }

// ContextWithFakeSpan returns a context carrying a fake OTel span with
// randomly-generated trace and span IDs.
// It returns the context plus the hex-encoded trace ID and span ID so callers
// can assert on them without hardcoding values.
func ContextWithFakeSpan(ctx context.Context) (context.Context, string, string) {
	var traceIDBytes [16]byte
	var spanIDBytes [8]byte
	_, _ = rand.Read(traceIDBytes[:])
	_, _ = rand.Read(spanIDBytes[:])

	traceID := trace.TraceID(traceIDBytes)
	spanID := trace.SpanID(spanIDBytes)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	return trace.ContextWithSpan(ctx, fakeSpan{sc: sc}), traceID.String(), spanID.String()
}
