package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/log"
)

func TestReplaceGlobals(t *testing.T) {
	t.Run("Should replace global logger and restore", func(t *testing.T) {
		// Initial setup: set global logger to a noop logger
		initialLogger := log.NewInMemory(log.LevelDebug)
		restoreFn := log.ReplaceGlobals(initialLogger)
		defer restoreFn()

		// Ensure the global logger is replaced
		assert.Equal(t, initialLogger, log.L())
	})

	// t.Run("Should return NoopLogger if no global logger is set", func(t *testing.T) {
	// 	// Reset to the no-op logger
	// 	restoreFn := log.ReplaceGlobals(nil)
	// 	defer restoreFn()

	// 	// Assert that the global logger is the no-op logger
	// 	assert.IsType(t, &log.NopLogger{}, log.L())
	// })
}
