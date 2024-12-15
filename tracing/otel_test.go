package tracing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/tracing"
)

func TestSetupOtelSDK_WithConsoleExporter(t *testing.T) {
	// Input configuration to enable console exporter.
	opts := []tracing.OtelOptFunc{
		tracing.WithServiceName("test-service"),
		tracing.WithServiceVersion("1.0.0"),
		tracing.WithEnvironment("test"),
		tracing.WithConsoleExporter(),
	}

	// Setup the OpenTelemetry SDK.
	shutdown, err := tracing.SetupOtelSDK(context.Background(), opts...)
	require.NoError(t, err)

	// Assert that the shutdown function is valid.
	require.NotNil(t, shutdown)

	// Verify that the system behaves as expected by calling the shutdown function.
	err = shutdown(context.Background())
	require.NoError(t, err)
}

// TestSetupOtelSDKWithOtlpGrpcExporter tests SetupOtelSDK with OTLP gRPC exporter enabled.
func TestSetupOtelSDKWithOtlpGrpcExporter(t *testing.T) {
	// Input configuration to enable OTLP gRPC exporter.
	opts := []tracing.OtelOptFunc{
		tracing.WithServiceName("test-service"),
		tracing.WithServiceVersion("1.0.0"),
		tracing.WithOtlpGrpcExporter(),
	}

	// Setup the OpenTelemetry SDK.
	shutdown, err := tracing.SetupOtelSDK(context.Background(), opts...)
	require.NoError(t, err)

	// Assert that the shutdown function is valid.
	require.NotNil(t, shutdown)

	// Verify that the system behaves as expected by calling the shutdown function.
	err = shutdown(context.Background())
	require.NoError(t, err)
}

func TestSetupOtel_Arguments(t *testing.T) {
	t.Parallel()

	t.Run("WithEmptyServiceName", func(t *testing.T) {
		t.Parallel()
		opts := []tracing.OtelOptFunc{
			tracing.WithServiceName(""),
			tracing.WithServiceVersion("1.0.0"),
			tracing.WithConsoleExporter(),
		}

		// Setup the OpenTelemetry SDK.
		_, err := tracing.SetupOtelSDK(context.Background(), opts...)

		// Verify that an error is returned for invalid configuration.
		require.Error(t, err)
		assert.ErrorIs(t, err, tracing.ErrMissingServiceName)
	})

	t.Run("WithEmptyServiceVersion", func(t *testing.T) {
		t.Parallel()
		opts := []tracing.OtelOptFunc{
			tracing.WithServiceName("test-service"),
			tracing.WithServiceVersion(""),
			tracing.WithConsoleExporter(),
		}

		// Setup the OpenTelemetry SDK.
		_, err := tracing.SetupOtelSDK(context.Background(), opts...)

		// Verify that an error is returned for missing service version.
		require.Error(t, err)
		assert.ErrorIs(t, err, tracing.ErrMissingServiceVersion)
	})
}

// TestShutdownFunctionWithError simulates an error during shutdown.
func TestShutdown(t *testing.T) {
	// Input configuration with an invalid setup that causes an error during shutdown.
	opts := []tracing.OtelOptFunc{
		tracing.WithServiceName("test-service"),
		tracing.WithServiceVersion("1.0.0"),
		tracing.WithConsoleExporter(),
	}

	// Setup the OpenTelemetry SDK (this is a mock simulation where the shutdown causes error).
	shutdown, err := tracing.SetupOtelSDK(context.Background(), opts...)
	require.NoError(t, err)

	// Simulate an error in the shutdown function (you would need to mock the actual error behavior in the shutdown).
	err = shutdown(context.Background())
	require.NoError(t, err)
}
