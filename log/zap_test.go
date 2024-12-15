package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/brpaz/lib-go/log"
)

// createTestLogger creates a ZapAdapter with an in-memory buffer for testing.
func createTestLogger(buf *bytes.Buffer, lvl zapcore.Level) *log.ZapAdapter {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(buf),
		zap.DebugLevel,
	)
	zapLogger := zap.New(core)
	return &log.ZapAdapter{
		ZapL:        zapLogger,
		AtomicLevel: zap.NewAtomicLevelAt(lvl),
	}
}

// assertContainsJSONField asserts that the JSON-encoded log output contains a specific field with a value.
func assertContainsJSONField(t *testing.T, jsonStr, key, expectedValue string) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	require.NoError(t, err, "failed to parse JSON log output")

	val, ok := data[key]
	require.True(t, ok, "key '%s' not found in log output: %s", key, jsonStr)
	assert.Equal(t, expectedValue, val, "unexpected value for key '%s'", key)
}

func TestZapAdapter_New(t *testing.T) {
	t.Parallel()

	t.Run("WithDefaultOptions", func(t *testing.T) {
		t.Parallel()
		logger, err := log.NewZap(log.DefaultZapLoggerOpts())

		require.NoError(t, err)
		assert.IsType(t, &log.ZapAdapter{}, logger)
		assert.Equal(t, log.LevelInfo, logger.GetLevel())
	})

	t.Run("WithInvalidProfile", func(t *testing.T) {
		t.Parallel()
		_, err := log.NewZap(log.ZapLoggerOpts{
			Profile: "invalid",
			Level:   log.LevelDebug,
			Format:  log.FormatJSON,
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, log.ErrInvalidProfile)
	})

	t.Run("WithInvalidFormat", func(t *testing.T) {
		t.Parallel()
		_, err := log.NewZap(log.ZapLoggerOpts{
			Format:  "invalid",
			Profile: log.ProfileDevelopment,
			Level:   log.LevelDebug,
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, log.ErrInvalidFormat)
	})
}

func TestZapAdapter_New_WithLevel(t *testing.T) {
	tests := []struct {
		name  string
		level log.Level
	}{
		{name: "Debug", level: log.LevelDebug},
		{name: "Info", level: log.LevelInfo},
		{name: "Warn", level: log.LevelWarn},
		{name: "Error", level: log.LevelError},
	}

	t.Parallel()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			logger, err := log.NewZap(log.ZapLoggerOpts{
				Level:   test.level,
				Profile: log.ProfileProduction,
				Format:  log.FormatJSON,
			})

			require.NoError(t, err)
			assert.IsType(t, &log.ZapAdapter{}, logger)
			assert.Equal(t, test.level, logger.GetLevel())
		})
	}
}

func TestZapAdapter_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf, zapcore.DebugLevel)

	ctx := context.Background()
	logger.Info(ctx, "info message", log.String("key", "value"))

	output := buf.String()

	assertContainsJSONField(t, output, "level", "info")
	assertContainsJSONField(t, output, "msg", "info message")
	assertContainsJSONField(t, output, "key", "value")
}

func TestZapAdapter_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf, zapcore.DebugLevel)

	ctx := context.Background()
	logger.Warn(ctx, "warning message", log.String("key", "value"))

	output := buf.String()

	assertContainsJSONField(t, output, "level", "warn")
	assertContainsJSONField(t, output, "msg", "warning message")
	assertContainsJSONField(t, output, "key", "value")
}

func TestZapAdapter_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf, zapcore.DebugLevel)

	ctx := context.Background()
	logger.Error(ctx, "error message", log.String("key", "value"))

	output := buf.String()

	assertContainsJSONField(t, output, "level", "error")
	assertContainsJSONField(t, output, "msg", "error message")
	assertContainsJSONField(t, output, "key", "value")
}

func TestZapAdapter_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf, zapcore.DebugLevel)

	ctx := context.Background()
	logger.Debug(ctx, "debug message", log.String("key", "value"))

	output := buf.String()

	assertContainsJSONField(t, output, "level", "debug")
	assertContainsJSONField(t, output, "msg", "debug message")
	assertContainsJSONField(t, output, "key", "value")
}

func TestZapAdapter_SetLevel(t *testing.T) {
	logger, err := log.NewZap(log.DefaultZapLoggerOpts())
	require.NoError(t, err)

	err = logger.SetLevel(log.LevelDebug)
	require.NoError(t, err)
	assert.Equal(t, log.LevelDebug, logger.GetLevel())
}

func TestZapAdapter_With(t *testing.T) {
	logger, err := log.NewZap(log.DefaultZapLoggerOpts())
	require.NoError(t, err)

	childLogger := logger.With(log.String("key", "value")).(*log.ZapAdapter)
	assert.NotEqual(t, logger, childLogger)
	assert.IsType(t, &log.ZapAdapter{}, childLogger)
	assert.Equal(t, log.LevelInfo, childLogger.GetLevel())
}

func TestZapAdapter_Sync(t *testing.T) {
	// Logger.Sync is gives an error when using stdout or stderr as the output.
	// So we are using a buffer to avoid the error.
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf, zapcore.DebugLevel)
	err := logger.Sync()
	require.NoError(t, err)
}

func TestZapAdapter_WithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf, zapcore.DebugLevel)

	loggerWithFields := logger.With(log.String("field1", "value1")).(*log.ZapAdapter)

	ctx := context.Background()
	loggerWithFields.Info(ctx, "info message")

	output := buf.String()
	assertContainsJSONField(t, output, "field1", "value1")
	assertContainsJSONField(t, output, "msg", "info message")
}
