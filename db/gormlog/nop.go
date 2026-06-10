package gormlog

import (
	"context"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// NopLogger is a [gormlogger.Interface] that discards all output.
type NopLogger struct{}

// NewNopLogger returns a NopLogger.
func NewNopLogger() *NopLogger { return &NopLogger{} }

func (NopLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface { return NopLogger{} }
func (NopLogger) Info(context.Context, string, ...any)             {}
func (NopLogger) Warn(context.Context, string, ...any)             {}
func (NopLogger) Error(context.Context, string, ...any)            {}
func (NopLogger) Trace(_ context.Context, _ time.Time, _ func() (string, int64), _ error) {
}
