package log

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// Constants for available profiles, adapters, and formats.
const (
	ProfileDevelopment = "dev"
	ProfileProduction  = "prod"
	AdapterZap         = "zap"
	AdapterNop         = "nop"
	AdapterInMemory    = "in-memory"
	FormatJSON         = "json"
	FormatLogFmt       = "logfmt"
)

// LoggerOpts specifies the options that can be used to configure the logger instance.
type LoggerOpts struct {
	Adapter string
	Level   Level
	Profile string
	Format  string
}

// LoggerOpt is a functional option type to configure the logger.
type LoggerOpt func(*LoggerOpts)

// Logger is a generic interface that defines the methods required by a logger.
type Logger interface {
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Debug(ctx context.Context, msg string, fields ...Field)
	SetLevel(level Level) error
	GetLevel() Level
	With(fields ...Field) Logger
	Sync() error
}

// Default values for LoggerOpts
var defaultOpts = LoggerOpts{
	Level:   LevelInfo,
	Profile: ProfileProduction,
	Format:  FormatJSON,
	Adapter: AdapterZap,
}

// Available valid options for profile, format, and adapter.
var (
	allowedProfiles = []string{ProfileDevelopment, ProfileProduction}
	allowedFormats  = []string{FormatJSON, FormatLogFmt}
	allowedAdapters = []string{AdapterZap, AdapterNop, AdapterInMemory}
)

var (
	ErrInvalidProfile = errors.New("invalid log profile")
	ErrInvalidFormat  = errors.New("invalid log format")
	ErrInvalidAdapter = errors.New("invalid log adapter")
)

// WithLevel sets the log level in the options.
func WithLevel(level Level) LoggerOpt {
	return func(o *LoggerOpts) {
		o.Level = level
	}
}

// WithProfile sets the logger profile in the options (e.g., dev, prod).
func WithProfile(profile string) LoggerOpt {
	return func(o *LoggerOpts) {
		o.Profile = profile
	}
}

// WithFormat specifies the log format to be used by the logger. Supported formats are json and console.
// This option will be only applicable for the Zap adapter.
func WithFormat(format string) LoggerOpt {
	return func(o *LoggerOpts) {
		o.Format = format
	}
}

// WithAdapter sets the logger adapter to be used by the logger.
func WithAdapter(adapter string) LoggerOpt {
	return func(o *LoggerOpts) {
		o.Adapter = adapter
	}
}

// Validate checks if the provided Logger options are valid.
func (opts *LoggerOpts) Validate() error {
	if !slices.Contains(allowedProfiles, opts.Profile) {
		return fmt.Errorf("%w: %s. Allowed values are %v", ErrInvalidProfile, opts.Profile, allowedProfiles)
	}
	if !slices.Contains(allowedFormats, opts.Format) {
		return fmt.Errorf("%w: %s. Allowed values are %v", ErrInvalidFormat, opts.Format, allowedFormats)
	}
	if !slices.Contains(allowedAdapters, opts.Adapter) {
		return fmt.Errorf("%w: %s. Allowed values are %v", ErrInvalidAdapter, opts.Adapter, allowedAdapters)
	}
	return nil
}

// New creates a new logger instance using the provided options.
// Example:
//
//	package main
//	import (
//	    "context"
//	    "github.com/brpaz/lib-go/log"
//	)
//	func main() {
//	    logger, err := log.New(
//	        log.WithAdapter("zap"),
//	        log.WithLevel(log.LevelDebug),
//	        log.WithProfile(log.ProfileProduction),
//	        log.WithFormat(log.FormatJSON),
//	    )
//	    if err != nil {
//	        panic(errors.Join(err, errors.New("failed to create logger")))
//	    }
//	    defer logger.Sync()
//	    logger.Info(context.Background(), "Application started", log.String("env", "production"))
//	}
func New(options ...LoggerOpt) (Logger, error) {
	// Apply options to default opts
	opts := defaultOpts
	for _, option := range options {
		option(&opts)
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create logger based on selected adapter
	switch opts.Adapter {
	case AdapterZap:
		return NewZap(ZapLoggerOpts{
			Level:   opts.Level,
			Profile: opts.Profile,
			Format:  opts.Format,
		})
	case AdapterNop:
		return NewNop(), nil
	case AdapterInMemory:
		return NewInMemory(opts.Level), nil
	default:
		return nil, ErrInvalidAdapter
	}
}
