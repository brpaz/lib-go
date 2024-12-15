package log

import "context"

// NopLogger is a no-op logger implementation.
type NopLogger struct{}

// NewNop returns a new instance of the no-op logger.
func NewNop() *NopLogger {
	return &NopLogger{}
}

// Info does nothing for the no-op logger.
func (l *NopLogger) Info(ctx context.Context, msg string, fields ...Field) {}

// Warn does nothing for the no-op logger.
func (l *NopLogger) Warn(ctx context.Context, msg string, fields ...Field) {}

// Error does nothing for the no-op logger.
func (l *NopLogger) Error(ctx context.Context, msg string, fields ...Field) {}

// Debug does nothing for the no-op logger.
func (l *NopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}

// SetLevel does nothing for the no-op logger.
func (l *NopLogger) SetLevel(level Level) error {
	return nil
}

// GetLevel always returns LevelInfo for the no-op logger.
func (l *NopLogger) GetLevel() Level {
	return LevelInfo
}

// With returns the same no-op logger instance.
func (l *NopLogger) With(fields ...Field) Logger {
	return l
}

// Sync does nothing for the no-op logger.
func (l *NopLogger) Sync() error {
	return nil
}
