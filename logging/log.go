package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// Format controls the log output encoding.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Middleware wraps an [slog.Handler] to inject additional attributes.
// Middlewares receive the context on every log call, making them suitable for
// extracting request-scoped values (e.g. user_id, tenant_id) from ctx.
type Middleware func(slog.Handler) slog.Handler

type config struct {
	environment string
	version     string
	revision    string
	format      Format
	level       string
	output      io.Writer
	middlewares []Middleware
	attrs       []slog.Attr
}

// Option configures a [Logger].
type Option func(*config)

// WithEnvironment sets the "environment" global attribute (e.g. "production").
func WithEnvironment(v string) Option {
	return func(c *config) { c.environment = v }
}

// WithVersion sets the "version" global attribute.
func WithVersion(v string) Option {
	return func(c *config) { c.version = v }
}

// WithRevision sets the "revision" global attribute (e.g. a git commit hash).
func WithRevision(v string) Option {
	return func(c *config) { c.revision = v }
}

// WithFormat sets the log output format.
func WithFormat(f Format) Option {
	return func(c *config) { c.format = f }
}

// WithLevel sets the initial log level as a string ("debug", "info", "warn", "error").
func WithLevel(l string) Option {
	return func(c *config) { c.level = l }
}

// WithOutput sets the writer for log output. Defaults to [os.Stdout].
func WithOutput(w io.Writer) Option {
	return func(c *config) { c.output = w }
}

// WithAttrs appends global attributes added to every log record.
func WithAttrs(attrs ...slog.Attr) Option {
	return func(c *config) { c.attrs = append(c.attrs, attrs...) }
}

// WithMiddleware appends one or more handler middlewares to the chain.
// Middlewares are applied in order after the built-in OTel handler, so they
// receive a context on every Handle call and can inject attrs derived from it.
func WithMiddleware(m ...Middleware) Option {
	return func(c *config) { c.middlewares = append(c.middlewares, m...) }
}

// Logger is a structured logger backed by [log/slog].
type Logger struct {
	inner    *slog.Logger
	levelVar *slog.LevelVar
}

// NewLogger constructs a [Logger] with the provided options.
// The logger includes an OTel handler that injects trace_id and span_id from
// the context on every log call.
func NewLogger(opts ...Option) (*Logger, error) {
	cfg := &config{
		format: FormatJSON,
		level:  "info",
		output: os.Stdout,
	}
	for _, o := range opts {
		o(cfg)
	}

	var lv slog.LevelVar
	if err := lv.UnmarshalText([]byte(cfg.level)); err != nil {
		return nil, fmt.Errorf("log: invalid level %q: %w", cfg.level, err)
	}

	handlerOpts := &slog.HandlerOptions{
		Level: &lv,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "timestamp"
			case slog.MessageKey:
				a.Key = "message"
			}
			return a
		},
	}

	var base slog.Handler
	if cfg.environment == "development" && cfg.format != FormatJSON {
		base = tint.NewHandler(cfg.output, &tint.Options{
			Level:       &lv,
			ReplaceAttr: handlerOpts.ReplaceAttr,
		})
	} else {
		switch cfg.format {
		case FormatJSON:
			base = slog.NewJSONHandler(cfg.output, handlerOpts)
		case FormatText:
			base = slog.NewTextHandler(cfg.output, handlerOpts)
		default:
			return nil, fmt.Errorf("log: unsupported format %q", cfg.format)
		}
	}

	var h slog.Handler = base
	for _, m := range cfg.middlewares {
		h = m(h)
	}

	globalAttrs := []any{
		slog.String("environment", cfg.environment),
		slog.String("version", cfg.version),
		slog.String("revision", cfg.revision),
	}
	for _, a := range cfg.attrs {
		globalAttrs = append(globalAttrs, a)
	}

	inner := slog.New(h).With(globalAttrs...)

	return &Logger{inner: inner, levelVar: &lv}, nil
}

// Info logs at INFO level.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.inner.InfoContext(ctx, msg, args...)
}

// Error logs at ERROR level.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.inner.ErrorContext(ctx, msg, args...)
}

// Warn logs at WARN level.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.inner.WarnContext(ctx, msg, args...)
}

// Debug logs at DEBUG level.
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.inner.DebugContext(ctx, msg, args...)
}

// With returns a new Logger pre-populated with the given attributes.
// The new logger shares the same [slog.LevelVar] as the parent, so a level
// change on either propagates to both.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{inner: l.inner.With(args...), levelVar: l.levelVar}
}

// LevelVar returns the runtime-mutable level variable.
func (l *Logger) LevelVar() *slog.LevelVar {
	return l.levelVar
}

// GetLevel returns the current log level.
func (l *Logger) GetLevel() slog.Level {
	return l.levelVar.Level()
}

// SetLevel changes the active log level at runtime.
func (l *Logger) SetLevel(level slog.Level) {
	l.levelVar.Set(level)
}

// Slog returns the underlying [slog.Logger] for interoperability.
func (l *Logger) Slog() *slog.Logger {
	return l.inner
}

// SetAsDefault installs this logger as the global [slog] default.
// After this call, package-level slog functions (slog.Info, slog.Error, …)
// and legacy log.Print calls route through this logger's handler.
// Call this once at the application wiring layer — never in library code.
func (l *Logger) SetAsDefault() {
	slog.SetDefault(l.inner)
}

// NewNoopLogger returns a Logger that discards all output.
// Useful in tests for components that require a Logger but produce no assertions on logs.
func NewNoopLogger() *Logger {
	return &Logger{
		inner:    slog.New(slog.DiscardHandler),
		levelVar: new(slog.LevelVar),
	}
}

func newNoop() *Logger { return NewNoopLogger() }
