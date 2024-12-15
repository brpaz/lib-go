package exporter

import (
	"context"

	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// NoopExporter is a no-op implementation of the SpanExporter interface.
type NoopExporter struct{}

// NewNoopExporter returns a new instance of the NoopExporter.
func NewNoopExporter() sdkTrace.SpanExporter {
	return &NoopExporter{}
}

func (NoopExporter) ExportSpans(ctx context.Context, spans []sdkTrace.ReadOnlySpan) error {
	return nil
}

func (NoopExporter) Shutdown(ctx context.Context) error {
	return nil
}
