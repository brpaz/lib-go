// Package gormlog provides [gormlogger.Interface] adapters for plugging
// structured loggers into GORM.
//
// # Adapters
//
// [New] wraps a [github.com/brpaz/lib-go/logging.Logger]. The GORM log
// level is derived automatically from the logger's current level, keeping
// query filtering consistent with the rest of the application. Slow-query and
// trace behaviour can be tuned with [WithSlowThreshold] and [WithIgnoreTrace].
//
// [NewNopLogger] discards all output. Useful in tests or when GORM logging
// should be suppressed without touching the slog configuration.
//
// # Adding adapters
//
// Any type that implements [gorm.io/gorm/logger.Interface] can be used as a
// GORM logger. To add support for another logging library (e.g. zap, logrus),
// add a new file in this package that implements the four interface methods:
// LogMode, Info, Warn, Error, and Trace.
package gormlog
