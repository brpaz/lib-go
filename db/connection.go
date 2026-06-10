package db

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	gormlogger "gorm.io/gorm/logger"

	"github.com/brpaz/lib-go/db/gormlog"
	"github.com/brpaz/lib-go/logging"
)

const (
	defaultMaxIdleConns    = 10
	defaultMaxOpenConns    = 100
	defaultConnMaxLifetime = 30 * time.Minute
)

// ConnOpts holds configuration for the database connection.
type ConnOpts struct {
	MaxIdleConns           int
	MaxOpenConns           int
	ConnMaxLifetime        time.Duration
	QueryLog               bool
	QueryLogShlowThreshold time.Duration
	Metrics                bool
	Tracing                bool
	DisableAutomaticPing   bool
	Logger                 *logging.Logger
}

// Validate checks the connection options for correctness.
func (o *ConnOpts) Validate() error {
	var errs []error
	if o.MaxIdleConns < 0 {
		errs = append(errs, errors.New("MaxIdleConns must be >= 0"))
	}
	if o.MaxOpenConns < 0 {
		errs = append(errs, errors.New("MaxOpenConns must be >= 0"))
	}
	if o.ConnMaxLifetime < 0 {
		errs = append(errs, errors.New("ConnMaxLifetime must be >= 0"))
	}
	return errors.Join(errs...)
}

// WithMaxIdleConns sets the maximum number of idle connections in the pool.
func WithMaxIdleConns(n int) func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.MaxIdleConns = n
	}
}

// WithMaxOpenConns sets the maximum number of open connections to the database.
func WithMaxOpenConns(n int) func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.MaxOpenConns = n
	}
}

// WithConnMaxLifetime sets the maximum amount of time a connection may be reused.
func WithConnMaxLifetime(d time.Duration) func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.ConnMaxLifetime = d
	}
}

// WithQueryLog enables GORM query logging via the configured logger.
func WithQueryLog() func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.QueryLog = true
	}
}

func WithQueryLogSlowThreshold(d time.Duration) func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.QueryLogShlowThreshold = d
	}
}

// WithMetrics enables GORM Prometheus metrics collection.
func WithMetrics() func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.Metrics = true
	}
}

// WithTracing enables OpenTelemetry tracing for GORM operations.
func WithTracing() func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.Tracing = true
	}
}

// WithAutomaticPing enables the automatic ping GORM performs after opening a connection,
// verifying the database is reachable before the first query.
func WithAutomaticPing() func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.DisableAutomaticPing = false
	}
}

// WithLogger sets the [logging.Logger] used for GORM query logging.
// Defaults to a no-op logger when not provided.
func WithLogger(logger *logging.Logger) func(*ConnOpts) {
	return func(opts *ConnOpts) {
		opts.Logger = logger
	}
}

// NewConnection opens a GORM connection using the provided dialector and options.
//
// The dialector selects the database driver (e.g. [gorm.io/driver/postgres.New]),
// keeping this package independent of any specific driver.
func NewConnection(dialector gorm.Dialector, opts ...func(*ConnOpts)) (*gorm.DB, error) {
	connOpts := &ConnOpts{
		MaxIdleConns:           defaultMaxIdleConns,
		MaxOpenConns:           defaultMaxOpenConns,
		ConnMaxLifetime:        defaultConnMaxLifetime,
		QueryLog:               false,
		QueryLogShlowThreshold: 200 * time.Millisecond,
		Metrics:                false,
		Tracing:                false,
		DisableAutomaticPing:   true,
		Logger:                 logging.NewNoopLogger(),
	}

	for _, opt := range opts {
		opt(connOpts)
	}

	if err := connOpts.Validate(); err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:               buildLogger(connOpts),
		DisableAutomaticPing: connOpts.DisableAutomaticPing,
	})
	if err != nil {
		return nil, fmt.Errorf("db: failed to open connection: %w", err)
	}

	// Configure the underlying *sql.DB connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db: failed to retrieve sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(connOpts.MaxIdleConns)
	sqlDB.SetMaxOpenConns(connOpts.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(connOpts.ConnMaxLifetime)

	// Optional: OpenTelemetry tracing plugin.
	if connOpts.Tracing {
		if err := db.Use(tracing.NewPlugin()); err != nil {
			return nil, fmt.Errorf("db: failed to register tracing plugin: %w", err)
		}
	}

	return db, nil
}

// buildLogger constructs the GORM logger from connection options.
// When query logging is disabled, GORM's built-in silent logger is used.
func buildLogger(opts *ConnOpts) gormlogger.Interface {
	if !opts.QueryLog {
		return gormlogger.Default.LogMode(gormlogger.Silent)
	}

	return gormlog.New(opts.Logger,
		gormlog.WithSlowThreshold(opts.QueryLogShlowThreshold),
	)
}
