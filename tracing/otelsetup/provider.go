package otelsetup

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
		return nil, fmt.Errorf("tracing/otelsetup: create resource: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("tracing/otelsetup: create exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
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
		return fmt.Errorf("tracing/otelsetup: shutdown: %w", err)
	}
	return nil
}

func newResource(cfg *options) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceVersion(cfg.serviceVersion),
		serviceRevisionKey.String(cfg.serviceRevision),
	}
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
}
