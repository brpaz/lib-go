package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan starts a new span using the named tracer from the global provider.
// tracerName should be the fully-qualified package path of the calling package
// (e.g. "github.com/acme/app/internal/orders").
func StartSpan(ctx context.Context, tracerName, spanName string) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, spanName)
}

// EndSpan records err on span (if non-nil) and ends it.
// Use with named returns and a defer closure:
//
//	func (s *Svc) Op(ctx context.Context) (_ Result, err error) {
//	    ctx, span := tracing.StartSpan(ctx, tracerName, "Op")
//	    defer func() { tracing.EndSpan(span, err) }()
func EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
