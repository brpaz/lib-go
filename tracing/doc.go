// Package tracing abstracts the OpenTelemetry SDK setup and provides a simple interface to interact with the OpenTelemetry API.
//
// Examples:
//
//	package main
//
//	import (
//		"context"
//		"log"
//
//	  	"github.com/brpaz/lib-go/tracing"
//
//	)
//
//	func main() {
//		// Setup the OpenTelemetry SDK with the console exporter.
//		shutdown, err := tracing.SetupOtelSDK(context.Background(), tracing.WithServiceName("test-service"), tracing.WithServiceVersion("1.0.0"), tracing.WithConsoleExporter())
//		if err != nil {
//			log.Fatalf("Error setting up OpenTelemetry SDK: %s", err)
//		}
//
//		// Gracefully shutdown the OpenTelemetry SDK.
//
//	defer func() {
//		_ = shutdown(context.Background())
//	}()
//
//		// Your application code here.
//	}
package tracing
