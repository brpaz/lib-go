package gormlog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	gormlogger "gorm.io/gorm/logger"

	"github.com/brpaz/lib-go/logging"
)

const defaultSlowThreshold = 200 * time.Millisecond

// Logger is a GORM [gormlogger.Interface] adapter backed by a [logging.Logger].
//
// The effective log level is derived from the logger at construction via
// gormLevelFromSlog: [logging.Logger.GetLevel] is mapped to a GORM level so
// that filtering stays consistent with the rest of the application.
//
// [Logger.LogMode] overrides the GORM level (e.g. to silence GORM output
// independently of the application logger) and returns a shallow clone.
//
// Level mapping:
//   - Normal queries are emitted at [slog.LevelInfo].
//   - Slow queries are emitted at [slog.LevelWarn].
//   - Errors are emitted at [slog.LevelError].
type Logger struct {
	logger        *logging.Logger
	logLevel      gormlogger.LogLevel
	slowThreshold time.Duration
	ignoreTrace   bool
}

// Option configures a [Logger].
type Option func(*Logger)

// WithSlowThreshold sets the duration above which a query is considered slow
// and logged as a warning. Defaults to 200ms. Set to 0 to disable slow-query logging.
func WithSlowThreshold(d time.Duration) Option {
	return func(l *Logger) {
		l.slowThreshold = d
	}
}

// WithIgnoreTrace disables SQL query tracing for non-error, non-slow queries.
// Errors and slow queries are still logged.
func WithIgnoreTrace() Option {
	return func(l *Logger) {
		l.ignoreTrace = true
	}
}

// New constructs a [Logger] backed by the given [logging.Logger].
// The initial GORM log level is derived from the logger's current level.
func New(l *logging.Logger, opts ...Option) *Logger {
	logger := &Logger{
		logger:        l,
		logLevel:      gormLevelFromSlog(l.GetLevel()),
		slowThreshold: defaultSlowThreshold,
	}
	for _, opt := range opts {
		opt(logger)
	}

	return logger
}

// gormLevelFromSlog maps a [slog.Level] to a GORM log level.
func gormLevelFromSlog(level slog.Level) gormlogger.LogLevel {
	switch {
	case level <= slog.LevelInfo:
		return gormlogger.Info
	case level <= slog.LevelWarn:
		return gormlogger.Warn
	case level <= slog.LevelError:
		return gormlogger.Error
	default:
		return gormlogger.Silent
	}
}

// LogMode returns a shallow copy of the logger with the GORM log level set to level.
// Pass [gormlogger.Silent] to suppress all output from this logger.
func (l *Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.logLevel = level
	return &clone
}

// Info logs an informational GORM message.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	if l.logLevel < gormlogger.Info {
		return
	}

	l.logger.Info(ctx, fmt.Sprintf(msg, args...))
}

// Warn logs a GORM warning message.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	if l.logLevel < gormlogger.Warn {
		return
	}

	l.logger.Warn(ctx, fmt.Sprintf(msg, args...))
}

// Error logs a GORM error message.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	if l.logLevel < gormlogger.Error {
		return
	}

	l.logger.Error(ctx, fmt.Sprintf(msg, args...))
}

// Trace logs a SQL query execution. It is called by GORM after every query.
//
// Behaviour:
//   - Errors (excluding ErrRecordNotFound) are logged at [slog.LevelError] when logLevel >= Error.
//   - Queries exceeding [slowThreshold] are logged at [slog.LevelWarn] when logLevel >= Warn.
//   - All other queries are logged at [slog.LevelInfo] when logLevel >= Info and ignoreTrace is unset.
func (l *Logger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (string, int64),
	err error,
) {
	if l.logLevel == gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	attrs := []slog.Attr{
		slog.Duration("elapsed", elapsed),
		slog.Int64("rows", rows),
		slog.String("sql", sql),
	}

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		if l.logLevel >= gormlogger.Error {
			l.logger.Error(ctx, "gorm query error",
				attrArgs(append(attrs, slog.String("error", err.Error())))...,
			)
		}

	case l.slowThreshold > 0 && elapsed > l.slowThreshold:
		if l.logLevel >= gormlogger.Warn {
			l.logger.Warn(ctx, "gorm slow query",
				attrArgs(append(attrs, slog.Duration("threshold", l.slowThreshold)))...,
			)
		}

	case !l.ignoreTrace:
		if l.logLevel >= gormlogger.Info {
			l.logger.Info(ctx, "gorm query", attrArgs(attrs)...)
		}
	}
}

// attrArgs converts attrs to a slice suitable for [logging.Logger]'s
// variadic key-value arguments.
func attrArgs(attrs []slog.Attr) []any {
	args := make([]any, len(attrs))
	for i, a := range attrs {
		args[i] = a
	}
	return args
}
