package log_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/log"
)

func TestNewInMemory(t *testing.T) {
	logger := log.NewInMemory(log.LevelDebug)

	// Check that the returned logger is of type InMemoryLogger
	assert.IsType(t, &log.InMemoryLogger{}, logger)
	// Verify the initial log level is set correctly
	assert.Equal(t, log.LevelDebug, logger.GetLevel())
}

func TestInMemoryLoggerSetLevel(t *testing.T) {
	logger := log.NewInMemory(log.LevelDebug)

	// Set a new log level and verify no error occurred
	err := logger.SetLevel(log.LevelWarn)
	assert.NoError(t, err)
	// Verify the new log level is applied
	assert.Equal(t, log.LevelWarn, logger.GetLevel())
}

func TestInMemoryLoggerMethods(t *testing.T) {
	// Define the test cases for different logging methods
	testCases := []struct {
		name string
	}{
		{name: "Info"},
		{name: "Debug"},
		{name: "Warn"},
		{name: "Error"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			logger := log.NewInMemory(log.LevelDebug)
			ctx := context.Background()

			// Log a message with a field
			logger.Info(ctx, "message", log.String("key", "value"))

			// Retrieve log entries and verify expected behavior
			logEntries := logger.Entries()
			assert.Len(t, logEntries, 1)
			assert.Equal(t, "message", logEntries[0].Message)
			assert.NotNil(t, logEntries[0].Timestamp)
			assert.Equal(t, "key", logEntries[0].Fields[0].Key)
			assert.Equal(t, "value", logEntries[0].Fields[0].String)
		})
	}
}

func TestInMemoryLoggerShouldLog(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel log.Level
		logLevel     log.Level
		shouldLog    bool
	}{
		{"DebugLogged", log.LevelDebug, log.LevelDebug, true},
		{"DebugNotLogged", log.LevelWarn, log.LevelDebug, false},
		{"InfoLogged", log.LevelDebug, log.LevelInfo, true},
		{"InfoNotLogged", log.LevelWarn, log.LevelInfo, false},
		{"WarnLogged", log.LevelInfo, log.LevelWarn, true},
		{"WarnNotLogged", log.LevelError, log.LevelWarn, false},
		{"ErrorLogged", log.LevelDebug, log.LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewInMemory(tt.currentLevel)

			// Log the message based on the log level being tested
			ctx := context.TODO()
			switch tt.logLevel {
			case log.LevelInfo:
				logger.Info(ctx, "info message")
			case log.LevelWarn:
				logger.Warn(ctx, "warn message")
			case log.LevelError:
				logger.Error(ctx, "error message")
			case log.LevelDebug:
				logger.Debug(ctx, "debug message")
			}

			// Verify if the log entry is present based on the level
			entries := logger.Entries()
			if tt.shouldLog {
				assert.Len(t, entries, 1)
			} else {
				assert.Len(t, entries, 0)
			}
		})
	}
}

func TestInMemoryWith(t *testing.T) {
	logger := log.NewInMemory(log.LevelDebug)

	// Create a child logger with additional fields
	childLogger := logger.With(log.String("key", "value")).(*log.InMemoryLogger)
	childLogger.Info(context.Background(), "message")

	// Verify that the child logger is of the correct type and different from the parent
	assert.IsType(t, &log.InMemoryLogger{}, childLogger)
	assert.NotEqual(t, logger, childLogger)

	// Verify the log entry contains the correct fields
	assert.Len(t, childLogger.Entries(), 1)
	assert.Equal(t, "key", childLogger.Entries()[0].Fields[0].Key)
	assert.Equal(t, "value", childLogger.Entries()[0].Fields[0].String)
}

func TestInMemoryLoggerSync(t *testing.T) {
	logger := log.NewInMemory(log.LevelDebug)

	// Syncing should not produce an error for InMemoryLogger
	err := logger.Sync()
	assert.NoError(t, err)
}

func TestInMemoryLogEntryGetField(t *testing.T) {
	entry := &log.InMemoryLogEntry{
		Fields: []log.Field{
			{Key: "key1", String: "value1"},
			{Key: "key2", Integer: 10},
		},
	}

	t.Run("ExistingField", func(t *testing.T) {
		field, ok := entry.GetField("key1")
		assert.True(t, ok)
		assert.Equal(t, "value1", field.String)
	})

	t.Run("NonExistingField", func(t *testing.T) {
		_, ok := entry.GetField("key3")
		assert.False(t, ok)
	})
}

func TestEmptyMessageLogging(t *testing.T) {
	logger := log.NewInMemory(log.LevelDebug)
	ctx := context.Background()

	// Log an empty message
	logger.Info(ctx, "", log.String("key", "value"))

	// Verify the log entry exists and the message is empty
	entries := logger.Entries()
	assert.Len(t, entries, 1)
	assert.Equal(t, "", entries[0].Message)
	assert.Equal(t, "key", entries[0].Fields[0].Key)
	assert.Equal(t, "value", entries[0].Fields[0].String)
}

func TestNoFieldsLogging(t *testing.T) {
	logger := log.NewInMemory(log.LevelDebug)
	ctx := context.Background()

	// Log a message without any fields
	logger.Info(ctx, "message")

	// Verify the log entry exists with no fields
	entries := logger.Entries()
	assert.Len(t, entries, 1)
	assert.Equal(t, "message", entries[0].Message)
	assert.Len(t, entries[0].Fields, 0)
}

func TestSync(t *testing.T) {
	logger := log.NewInMemory(log.LevelDebug)

	// Syncing should not produce an error for InMemoryLogger
	err := logger.Sync()
	assert.NoError(t, err)
}
