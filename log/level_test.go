package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/log"
)

func TestLevelFromString_ValidLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected log.Level
	}{
		{name: "Lowercase Debug", input: "debug", expected: log.LevelDebug},
		{name: "Lowercase Info", input: "info", expected: log.LevelInfo},
		{name: "Lowercase Warn", input: "warn", expected: log.LevelWarn},
		{name: "Lowercase Error", input: "error", expected: log.LevelError},
		{name: "Uppercase Debug", input: "DEBUG", expected: log.LevelDebug},
		{name: "Uppercase Info", input: "INFO", expected: log.LevelInfo},
		{name: "Uppercase Warn", input: "WARN", expected: log.LevelWarn},
		{name: "Uppercase Error", input: "ERROR", expected: log.LevelError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			level, err := log.LevelFromString(test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, level)
		})
	}
}

func TestLevelFromString_InvalidLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "Empty String", input: ""},
		{name: "Completely Invalid", input: "invalid"},
		{name: "Partially Matching", input: "Debug123"},
		{name: "Leading Whitespace", input: " warn"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			level, err := log.LevelFromString(test.input)
			assert.Error(t, err)
			assert.Equal(t, log.LevelInfo, level)
			assert.Equal(t, log.ErrInvalidLevel, err)
		})
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    log.Level
		expected string
	}{
		{
			level:    log.LevelDebug,
			expected: "debug",
		},
		{
			level:    log.LevelInfo,
			expected: "info",
		},
		{
			level:    log.LevelWarn,
			expected: "warn",
		},
		{
			level:    log.LevelError,
			expected: "error",
		},
		{
			level:    log.Level(100),
			expected: "unknown",
		},
		{
			level:    log.Level(-1),
			expected: "unknown",
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}
