package tracing

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	libExporter "github.com/brpaz/lib-go/tracing/exporter"
)

var (
	ErrMissingServiceName    = errors.New("service name is required")
	ErrMissingServiceVersion = errors.New("service version is required")
)

// OtelOpts specifies the options for configuring the OpenTelemetry SDK.
type OtelOpts struct {
	ServiceName      string
	ServiceVersion   string
	Envrionment      string
	ConsoleExporter  bool
	OtlpGrpcExporter bool
}

// OtelOptFunc is a functional option type for configuring OtelOpts.
type OtelOptFunc func(*OtelOpts)

// defaultOtelOpts are the default OpenTelemetry options.
var defaultOtelOpts = OtelOpts{}

// Validate checks the validity of the options.
func (o *OtelOpts) Validate() error {
	if o.ServiceName == "" {
		return ErrMissingServiceName
	}
	if o.ServiceVersion == "" {
		return ErrMissingServiceVersion
	}

	return nil
}

// WithServiceName configures the service name.
func WithServiceName(serviceName string) OtelOptFunc {
	return func(o *OtelOpts) {
		o.ServiceName = serviceName
	}
}

// WithServiceVersion configures the service version.
func WithServiceVersion(serviceVersion string) OtelOptFunc {
	return func(o *OtelOpts) {
		o.ServiceVersion = serviceVersion
	}
}

func WithEnvironment(env string) OtelOptFunc {
	return func(o *OtelOpts) {
		o.Envrionment = env
	}
}

// WithConsoleExporter enables the console exporter.
func WithConsoleExporter() OtelOptFunc {
	return func(o *OtelOpts) {
		o.ConsoleExporter = true
	}
}

// WithOtlpGrpcExporter enables the OTLP gRPC exporter and configures the endpoint.
func WithOtlpGrpcExporter() OtelOptFunc {
	return func(o *OtelOpts) {
		o.OtlpGrpcExporter = true
	}
}

// SetupOtelSDK bootstraps the OpenTelemetry SDK and returns a shutdown function.
func SetupOtelSDK(ctx context.Context, options ...OtelOptFunc) (shutdown func(context.Context) error, err error) {
	opts := defaultOtelOpts
	for _, opt := range options {
		opt(&opts)
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid OpenTelemetry configuration: %w", err)
	}

	var shutdownFuncs []func(context.Context) error
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator
	otel.SetTextMapPropagator(newPropagator())

	// Set up trace provider
	tracerProvider, err := newTraceProvider(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Add shutdown function for tracer provider
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return shutdown, nil
}

// newPropagator sets up the trace context and baggage propagators.
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTraceProvider creates a new tracer provider with the appropriate exporter.
func newTraceProvider(ctx context.Context, opts OtelOpts) (*sdkTrace.TracerProvider, error) {
	traceExporter, err := newTraceExporter(ctx, opts)
	if err != nil {
		return nil, err
	}

	res, err := newResource(ctx, opts)
	if err != nil {
		return nil, err
	}

	traceProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithResource(res),
		sdkTrace.WithSampler(sdkTrace.ParentBased(sdkTrace.TraceIDRatioBased(0.8))),
		sdkTrace.WithBatcher(traceExporter),
	)
	return traceProvider, nil
}

// newResource creates a new OpenTelemetry resource with service info and telemetry metadata.
func newResource(ctx context.Context, opts OtelOpts) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithContainer(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(opts.ServiceName),
			semconv.ServiceVersionKey.String(opts.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(opts.Envrionment),
		),
	)
}

// newTraceExporter returns the appropriate trace exporter based on configuration.
func newTraceExporter(ctx context.Context, opts OtelOpts) (sdkTrace.SpanExporter, error) {
	if opts.ConsoleExporter {
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	}

	if opts.OtlpGrpcExporter {
		otlpClient := otlptracegrpc.NewClient()
		return otlptrace.New(ctx, otlpClient)
	}

	return libExporter.NewNoopExporter(), nil
}
