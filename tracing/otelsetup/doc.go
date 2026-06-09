// Package setup initialises an OpenTelemetry TracerProvider backed by an
// OTLP/gRPC exporter. It imports the full OTel SDK and a gRPC exporter, so
// consumers that only need span helpers should import
// [github.com/brpaz/lib-go/tracing] instead.
package otelsetup
