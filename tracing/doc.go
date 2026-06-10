// Package tracing provides OpenTelemetry distributed tracing helpers:
// a [Provider] that wires up an OTLP/gRPC TracerProvider, and span helpers
// ([StartSpan], [EndSpan]).
//
// For OpenTelemetry trace/span injection into logs, see
// [github.com/brpaz/lib-go/logging.OtelMiddleware].
//
// # Provider setup
//
//	provider, err := tracing.New(ctx,
//		tracing.WithServiceVersion("1.2.3"),
//		tracing.WithServiceRevision("abc1234"),
//	)
//	if err != nil {
//		return err
//	}
//	defer provider.Shutdown(ctx)
//
// The provider is registered as the global OTel tracer provider on creation,
// so instrumentation libraries (otelhttp, GORM plugin, etc.) work without
// explicit wiring after [New] returns. Endpoint, protocol and service name
// are read from the standard OTel env vars (OTEL_EXPORTER_OTLP_ENDPOINT,
// OTEL_EXPORTER_OTLP_PROTOCOL, OTEL_SERVICE_NAME).
//
// # Span helpers
//
//	func (s *Svc) Op(ctx context.Context) (_ Result, err error) {
//		ctx, span := tracing.StartSpan(ctx, tracerName, "Op")
//		defer func() { tracing.EndSpan(span, err) }()
//		...
//	}
//
// For HTTP middleware and test helpers, see the [tracing/middleware] and
// [tracing/tracingtest] sub-packages.
package tracing
