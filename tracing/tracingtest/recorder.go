package tracingtest

import (
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Recorder captures all spans emitted during a test.
// Install it with [NewRecorder] and call Cleanup when the test ends.
type Recorder struct {
	*tracetest.SpanRecorder
	cleanup func()
}

// NewRecorder installs an in-memory OTel provider that records all spans.
// The previous global provider is restored on Cleanup.
//
//	rec := tracingtest.NewRecorder()
//	t.Cleanup(rec.Cleanup)
func NewRecorder() *Recorder {
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	return &Recorder{
		SpanRecorder: rec,
		cleanup:      func() { otel.SetTracerProvider(prev) },
	}
}

// Cleanup restores the global tracer provider to the value it had before
// [NewRecorder] was called.
func (r *Recorder) Cleanup() { r.cleanup() }
