package log

import (
	"sync"
	"sync/atomic"
)

var (
	// Defines a global logger instance
	globalLogger   atomic.Value
	globalLoggerMu sync.Mutex
)

// ReplaceGlobals replaces the global Logger and returns a function to restore
// the original values. It's safe for concurrent use.
func ReplaceGlobals(logger Logger) func() {
	globalLoggerMu.Lock()
	defer globalLoggerMu.Unlock()

	// If the logger is nil, store the NoopLogger instead
	if logger == nil {
		logger = NewNop() // Default to a no-op logger
	}

	// Get the previous logger value
	prev := globalLogger.Load()

	// Store the new logger value
	globalLogger.Store(logger)

	// If the previous logger is nil, return a no-op function
	if prev == nil {
		return func() {}
	}

	// Restore the previous logger value
	return func() {
		globalLogger.Store(prev)
	}
}

// L retrieves the global logger instance. If no logger is set, returns a no-op logger.
func L() Logger {
	// Load the current logger instance from the atomic value
	logger, ok := globalLogger.Load().(Logger)
	if !ok {
		// If no valid logger is found, return a no-op logger
		return NewNop()
	}
	return logger
}
