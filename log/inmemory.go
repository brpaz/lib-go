package log

import (
	"context"
	"time"
)

// InMemoryLogger stores log entries in memory for testing purposes.
type InMemoryLogger struct {
	level   Level
	fields  []Field
	entries []InMemoryLogEntry
}

// InMemoryLogEntry represents a single log entry.
type InMemoryLogEntry struct {
	Timestamp time.Time
	Message   string
	Level     string
	Fields    []Field
}

// GetField retrieves a field by its key from the log entry.
func (e *InMemoryLogEntry) GetField(key string) (Field, bool) {
	for _, f := range e.Fields {
		if f.Key == key {
			return f, true
		}
	}
	return Field{}, false
}

// NewInMemory creates and returns a new instance of InMemoryLogger with the provided log level.
// Usage:
//
//	package main
//	import (
//		"context"
//		"github.com/brpaz/lib-go/log"
//	)
//
//	func main() {
//	   logger := log.NewInMemory(log.LevelInfo)
//	   logger.Info(context.Background(), "Application started")
//	   entries := logger.Entries()
//	   for _, entry := range entries {
//	       fmt.Println(entry.Message)
//	   }
//	}
func NewInMemory(level Level) *InMemoryLogger {
	return &InMemoryLogger{
		level:   level,
		fields:  nil, // No global fields by default
		entries: nil, // Start with no entries
	}
}

// log processes log messages by appending log entries if the log level permits.
func (l *InMemoryLogger) log(ctx context.Context, lvl Level, message string, fields ...Field) {
	if lvl < l.level {
		return // Skip logging if the log level is below the configured level
	}

	// Merge global logger fields with the specific fields for this log entry
	logFields := append(append(l.fields, fields...), ExtractTraceIDFieldsFromContext(ctx)...)

	// Store the log entry
	l.entries = append(l.entries, InMemoryLogEntry{
		Timestamp: time.Now(),
		Message:   message,
		Fields:    logFields,
		Level:     lvl.String(),
	})
}

// Info logs an informational message.
func (l *InMemoryLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelInfo, msg, fields...)
}

// Warn logs a warning message.
func (l *InMemoryLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelWarn, msg, fields...)
}

// Error logs an error message.
func (l *InMemoryLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelError, msg, fields...)
}

// Debug logs a debug message.
func (l *InMemoryLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelDebug, msg, fields...)
}

// SetLevel updates the logger's log level.
func (l *InMemoryLogger) SetLevel(level Level) error {
	l.level = level
	return nil
}

// GetLevel retrieves the current log level of the logger.
func (l *InMemoryLogger) GetLevel() Level {
	return l.level
}

// Entries returns all the log entries stored by the logger.
func (l *InMemoryLogger) Entries() []InMemoryLogEntry {
	return l.entries
}

// With returns a new logger instance with the additional fields.
func (l *InMemoryLogger) With(fields ...Field) Logger {
	return &InMemoryLogger{
		level:   l.level,
		fields:  append(l.fields, fields...),
		entries: l.entries,
	}
}

// Sync simulates synchronizing the logger (no-op for InMemoryLogger).
func (l *InMemoryLogger) Sync() error {
	return nil
}
