package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultPort              = 8080
	defaultReadTimeout       = 5 * time.Second
	defaultReadHeaderTimeout = 2 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
)

// Server wraps net/http.Server with lifecycle helpers.
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

// Option configures a Server.
type Option func(*options)

type options struct {
	port              int
	handler           http.Handler
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	shutdownTimeout   time.Duration
}

// WithPort sets the listening port.
func WithPort(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

// WithHandler sets the HTTP handler for the server.
// Pass a fully configured Router here; the server owns no routing logic.
func WithHandler(h http.Handler) Option {
	return func(o *options) {
		o.handler = h
	}
}

// WithReadTimeout sets the HTTP read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(o *options) {
		o.readTimeout = d
	}
}

// WithReadHeaderTimeout sets the maximum duration for reading request headers.
// Setting this protects against slow-header attacks (e.g. Slowloris).
func WithReadHeaderTimeout(d time.Duration) Option {
	return func(o *options) {
		o.readHeaderTimeout = d
	}
}

// WithWriteTimeout sets the HTTP write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(o *options) {
		o.writeTimeout = d
	}
}

// WithIdleTimeout sets the HTTP idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(o *options) {
		o.idleTimeout = d
	}
}

// WithShutdownTimeout sets how long Run waits for active connections to drain
// during a graceful shutdown before giving up.
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *options) {
		o.shutdownTimeout = d
	}
}

// validate checks that required options are set.
func (o *options) validate() error {
	if o.handler == nil {
		return errors.New("handler is required: use WithHandler to provide a configured Router")
	}
	if o.port <= 0 || o.port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", o.port)
	}
	return nil
}

// New creates a Server with the given options.
// Returns an error if required options are missing or invalid.
func New(opts ...Option) (*Server, error) {
	o := &options{
		port:              defaultPort,
		readTimeout:       defaultReadTimeout,
		readHeaderTimeout: defaultReadHeaderTimeout,
		writeTimeout:      defaultWriteTimeout,
		idleTimeout:       defaultIdleTimeout,
		shutdownTimeout:   defaultShutdownTimeout,
	}
	for _, opt := range opts {
		opt(o)
	}

	if err := o.validate(); err != nil {
		return nil, fmt.Errorf("server: %w", err)
	}

	s := &Server{
		shutdownTimeout: o.shutdownTimeout,
	}
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", o.port),
		Handler:           o.handler,
		ReadTimeout:       o.readTimeout,
		ReadHeaderTimeout: o.readHeaderTimeout,
		WriteTimeout:      o.writeTimeout,
		IdleTimeout:       o.idleTimeout,
	}

	return s, nil
}

// Addr returns the server's listen address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// Start begins serving HTTP requests. Blocks until the server stops.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully drains active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Run starts the server and blocks until ctx is cancelled, then performs a
// graceful shutdown bounded by the configured shutdown timeout (see
// WithShutdownTimeout). It returns nil on a clean shutdown, or the first
// error encountered from either Start or Shutdown.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server: shutdown: %w", err)
	}

	return <-errCh
}
