package log

import (
	"context"
	"errors"
	"reflect"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	gorml "gorm.io/gorm/logger"

	"github.com/brpaz/lib-go/log"
)

// GormLogger is a logger adapter between the application's Logger and GORM's logger interface.
type GormLogger struct {
	logger                    log.Logger
	logLevel                  logger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
	silent                    bool
}

// GormLoggerOpts contains the options for configuring a GormLogger.
type GormLoggerOpts struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	Logger                    log.Logger
	Silent                    bool
}

// GormLoggerOptionFunc is a function that modifies GormLoggerOpts.
type GormLoggerOptionFunc func(*GormLoggerOpts)

// Default options for the GormLogger.
var defaultLoggerOpts = GormLoggerOpts{
	SlowThreshold:             100 * time.Millisecond,
	IgnoreRecordNotFoundError: true,
	Silent:                    false,
}

// WithSlowThreshold sets the slow threshold for SQL queries.
func WithSlowThreshold(threshold time.Duration) GormLoggerOptionFunc {
	return func(opts *GormLoggerOpts) {
		opts.SlowThreshold = threshold
	}
}

// WithIgnoreRecordNotFoundError sets whether to ignore RecordNotFoundError errors.
func WithIgnoreRecordNotFoundError(ignore bool) GormLoggerOptionFunc {
	return func(opts *GormLoggerOpts) {
		opts.IgnoreRecordNotFoundError = ignore
	}
}

// WithSilent sets whether the logger should be silent.
func WithSilent(silent bool) GormLoggerOptionFunc {
	return func(opts *GormLoggerOpts) {
		opts.Silent = silent
	}
}

// NewGormLogger creates a new GormLogger instance with the provided options.
func NewGormLogger(logger log.Logger, opts ...GormLoggerOptionFunc) *GormLogger {
	loggerOpts := defaultLoggerOpts

	for _, opt := range opts {
		opt(&loggerOpts)
	}

	return &GormLogger{
		logger:                    logger,
		logLevel:                  mapLogLevel(logger.GetLevel()),
		slowThreshold:             loggerOpts.SlowThreshold,
		ignoreRecordNotFoundError: loggerOpts.IgnoreRecordNotFoundError,
		silent:                    loggerOpts.Silent,
	}
}

// LogLevel returns the current logging level.
func (gl *GormLogger) LogLevel() gorml.LogLevel {
	return gl.logLevel
}

// LogMode sets the logging level for a specific instance of GormLogger.
func (gl *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *gl
	newLogger.logLevel = level
	return &newLogger
}

// Info logs informational messages if the log level permits it.
func (gl *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if gl.silent {
		return
	}

	if gl.logLevel >= logger.Info {
		gl.logger.Info(ctx, msg, convertArgsToFields(args)...)
	}
}

// Warn logs warning messages if the log level permits it.
func (gl *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if gl.silent {
		return
	}

	if gl.logLevel >= logger.Warn {
		gl.logger.Warn(ctx, msg, convertArgsToFields(args)...)
	}
}

// Error logs error messages if the log level permits it.
func (gl *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if gl.silent {
		return
	}

	if gl.logLevel >= logger.Error {
		gl.logger.Error(ctx, msg, convertArgsToFields(args)...)
	}
}

// Trace logs SQL query details, including execution time, errors, and affected rows.
func (gl *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if gl.silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && (!gl.ignoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		gl.logger.Error(ctx, "SQL execution error",
			log.Error(err),
			log.Duration("elapsed", elapsed),
			log.Int64("rows", rows),
			log.String("sql", sql),
		)
	case elapsed > gl.slowThreshold && gl.slowThreshold > 0:
		gl.logger.Warn(ctx, "Slow SQL query",
			log.Duration("elapsed", elapsed),
			log.Int64("rows", rows),
			log.String("sql", sql),
		)
	default:
		gl.logger.Debug(ctx, "SQL query trace",
			log.Duration("elapsed", elapsed),
			log.Int64("rows", rows),
			log.String("sql", sql),
		)
	}
}

// mapLogLevel maps the application's log levels to Gorm's log levels.
func mapLogLevel(level log.Level) gorml.LogLevel {
	switch level {
	case log.LevelInfo:
		return gorml.Info
	case log.LevelWarn:
		return gorml.Warn
	case log.LevelError:
		return gorml.Error
	case log.LevelDebug:
		return gorml.Info
	default:
		return gorml.Info
	}
}

// helper function to convert arguments to log fields.
func convertArgsToFields(args []interface{}) []log.Field {
	fields := make([]log.Field, len(args))
	for i, arg := range args {
		t := reflect.TypeOf(arg)
		switch t.Kind() {
		case reflect.String:
			fields[i] = log.String(t.Name(), arg.(string))
		case reflect.Int:
			fields[i] = log.Int(t.Name(), arg.(int))
		case reflect.Int64:
			fields[i] = log.Int64(t.Name(), arg.(int64))
		case reflect.Float64:
			fields[i] = log.Float64(t.Name(), arg.(float64))
		case reflect.Bool:
			fields[i] = log.Bool(t.Name(), arg.(bool))
		default:
			fields[i] = log.String(t.Name(), arg.(string))
		}
	}
	return fields
}
