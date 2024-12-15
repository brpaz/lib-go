package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/log"
)

func createNopLogger() *log.NopLogger {
	return log.NewNop()
}

func TestNopLogger_Info(t *testing.T) {
	logger := createNopLogger()
	logger.Info(context.Background(), "info message", log.String("key", "value"))
	// No assertions as this is a no-op logger
}

func TestNopLogger_Warn(t *testing.T) {
	logger := createNopLogger()
	logger.Warn(context.Background(), "warn message", log.String("key", "value"))
	// No assertions as this is a no-op logger
}

func TestNopLogger_Error(t *testing.T) {
	logger := createNopLogger()
	logger.Error(context.Background(), "error message", log.String("key", "value"))
	// No assertions as this is a no-op logger
}

func TestNopLogger_Debug(t *testing.T) {
	logger := createNopLogger()
	logger.Debug(context.Background(), "debug message", log.String("key", "value"))
	// No assertions as this is a no-op logger
}

func TestNopLogger_SetLevel(t *testing.T) {
	logger := createNopLogger()
	err := logger.SetLevel(log.LevelDebug)
	assert.NoError(t, err)
}

func TestNopLogger_GetLevel(t *testing.T) {
	logger := createNopLogger()
	level := logger.GetLevel()
	assert.Equal(t, log.LevelInfo, level)
}

func TestNopLogger_With(t *testing.T) {
	logger := createNopLogger()
	newLogger := logger.With(log.String("key", "value"))
	assert.Equal(t, logger, newLogger)
}

func TestNopLogger_Sync(t *testing.T) {
	logger := createNopLogger()
	err := logger.Sync()
	assert.NoError(t, err)
}
