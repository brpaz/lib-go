package log

import (
	"context"
	"fmt"
	"slices"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapAdapter is a wrapper around zap.Logger that implements the Logger interface.
type ZapAdapter struct {
	ZapL        *zap.Logger
	AtomicLevel zap.AtomicLevel
}

// ZapLoggerOpts specifies options to configure the Zap logger instance.
type ZapLoggerOpts struct {
	Level   Level
	Profile string
	Format  string
}

func (o ZapLoggerOpts) Validate() error {
	if !slices.Contains(allowedProfiles, o.Profile) {
		return fmt.Errorf("%w: %s", ErrInvalidProfile, o.Profile)
	}

	if !slices.Contains(allowedFormats, o.Format) {
		return fmt.Errorf("%w: %s", ErrInvalidFormat, o.Format)
	}

	return nil
}

// defaultZapLoggerOpts defines the default options for the Zap logger.
var defaultZapLoggerOpts = ZapLoggerOpts{
	Level:   LevelInfo,
	Profile: ProfileProduction,
	Format:  FormatJSON,
}

// DefaultZapLoggerOpts returns a copy of the default logger options.
func DefaultZapLoggerOpts() ZapLoggerOpts {
	return defaultZapLoggerOpts
}

// NewZap creates a new Zap logger instance with the provided options.
func NewZap(opts ZapLoggerOpts) (*ZapAdapter, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	zapLogger, atomicLevel, err := createZapLogger(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}
	return &ZapAdapter{
		ZapL:        zapLogger,
		AtomicLevel: atomicLevel,
	}, nil
}

// log handles logging at various levels.
func (l *ZapAdapter) log(ctx context.Context, lvl zapcore.Level, msg string, fields ...Field) {
	fields = append(fields, ExtractTraceIDFieldsFromContext(ctx)...)
	l.ZapL.Log(lvl, msg, fields...)
}

// Info logs a message at the info level.
func (l *ZapAdapter) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zap.InfoLevel, msg, fields...)
}

// Warn logs a message at the warning level.
func (l *ZapAdapter) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zap.WarnLevel, msg, fields...)
}

// Error logs a message at the error level.
func (l *ZapAdapter) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zap.ErrorLevel, msg, fields...)
}

// Debug logs a message at the debug level.
func (l *ZapAdapter) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zap.DebugLevel, msg, fields...)
}

// SetLevel updates the logging level dynamically.
func (l *ZapAdapter) SetLevel(level Level) error {
	l.AtomicLevel.SetLevel(toZapLevel(level))
	return nil
}

// GetLevel retrieves the current logging level.
func (l *ZapAdapter) GetLevel() Level {
	return fromZapLevel(l.AtomicLevel.Level())
}

// With creates a new logger instance with additional fields.
func (l *ZapAdapter) With(fields ...Field) Logger {
	childLogger := l.ZapL.With(fields...)
	return &ZapAdapter{
		ZapL:        childLogger,
		AtomicLevel: l.AtomicLevel,
	}
}

// Sync flushes any buffered log entries.
func (l *ZapAdapter) Sync() error {
	return l.ZapL.Sync()
}

// createZapLogger initializes a Zap logger with the given options.
func createZapLogger(opts ZapLoggerOpts) (*zap.Logger, zap.AtomicLevel, error) {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(toZapLevel(opts.Level))

	config, err := getZapConfig(opts)
	if err != nil {
		return nil, atomicLevel, fmt.Errorf("failed to build zap config: %w", err)
	}
	logger, err := config.Build(zap.AddCallerSkip(2))
	return logger, atomicLevel, err
}

// getZapConfig returns the appropriate zap.Config for the given profile and options.
func getZapConfig(opts ZapLoggerOpts) (zap.Config, error) {
	var config zap.Config
	switch opts.Profile {
	case ProfileDevelopment:
		config = zap.NewDevelopmentConfig()
	default:
		config = zap.NewProductionConfig()
	}

	config.Encoding = getZapEncoding(opts.Format)
	config.EncoderConfig = getZapEncoderConfig()
	return config, nil
}

// getZapEncoding returns the encoding type based on format.
func getZapEncoding(format string) string {
	if format == FormatLogFmt {
		return "console"
	}
	return "json"
}

// getZapEncoderConfig defines common encoder configurations.
func getZapEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
}

// toZapLevel converts a Level to a zapcore.Level.
func toZapLevel(level Level) zapcore.Level {
	switch level {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// fromZapLevel converts a zapcore.Level to a Level.
func fromZapLevel(level zapcore.Level) Level {
	switch level {
	case zapcore.DebugLevel:
		return LevelDebug
	case zapcore.InfoLevel:
		return LevelInfo
	case zapcore.WarnLevel:
		return LevelWarn
	case zapcore.ErrorLevel:
		return LevelError
	default:
		return LevelInfo
	}
}
