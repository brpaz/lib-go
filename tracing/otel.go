package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

// serviceRevisionKey is the resource attribute key for the VCS revision (git SHA).
// Not yet part of stable semconv.
var serviceRevisionKey = attribute.Key("service.revision")

type options struct {
	serviceVersion  string
	serviceRevision string
	sampler         sdktrace.Sampler
	attrs           []attribute.KeyValue
	errHandler      func(err error)
}

// Option configures a [Provider].
type Option func(*options)

// WithServiceVersion sets the service.version resource attribute.
func WithServiceVersion(version string) Option {
	return func(o *options) { o.serviceVersion = version }
}

// WithServiceRevision sets the service.revision resource attribute (e.g. a git SHA).
func WithServiceRevision(revision string) Option {
	return func(o *options) { o.serviceRevision = revision }
}

// WithSampler sets the trace sampler. By default, the SDK's default sampler
// ([sdktrace.ParentBased] wrapping [sdktrace.AlwaysSample]) is used, which
// samples every trace.
func WithSampler(sampler sdktrace.Sampler) Option {
	return func(o *options) { o.sampler = sampler }
}

// WithAttributes adds extra resource attributes (e.g. deployment.environment),
// merged with the service.version and service.revision attributes.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return func(o *options) { o.attrs = attrs }
}

// WithErrorHandler registers fn to receive errors reported by the OTel SDK
// (e.g. export failures). By default, these are written to stderr.
func WithErrorHandler(fn func(err error)) Option {
	return func(o *options) { o.errHandler = fn }
}

// Provider wraps an OTel TracerProvider and provides lifecycle management.
// It is registered as the global OTel tracer provider on creation.
type Provider struct {
	tp       trace.TracerProvider
	shutdown func(context.Context) error
}

// New initialises an OTel [Provider] and registers it as the global tracer provider.
// Endpoint and protocol are read from standard OTel env vars
// (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_PROTOCOL).
// service.name is read from OTEL_SERVICE_NAME.
func New(ctx context.Context, opts ...Option) (*Provider, error) {
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}

	res, err := newResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("tracing: create resource: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("tracing: create exporter: %w", err)
	}

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	}
	if cfg.sampler != nil {
		tpOpts = append(tpOpts, sdktrace.WithSampler(cfg.sampler))
	}

	if cfg.errHandler != nil {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(cfg.errHandler))
	}

	tp := sdktrace.NewTracerProvider(tpOpts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{tp: tp, shutdown: tp.Shutdown}, nil
}

// Shutdown flushes buffered spans and releases provider resources.
// Defer immediately after a successful [New] call.
func (p *Provider) Shutdown(ctx context.Context) error {
	if err := p.shutdown(ctx); err != nil {
		return fmt.Errorf("tracing: shutdown: %w", err)
	}
	return nil
}

func newResource(cfg *options) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceVersion(cfg.serviceVersion),
		serviceRevisionKey.String(cfg.serviceRevision),
	}
	attrs = append(attrs, cfg.attrs...)
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
}
