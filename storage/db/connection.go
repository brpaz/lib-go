package db

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	gorml "gorm.io/gorm/logger"

	"github.com/brpaz/lib-go/log"
	dbLogger "github.com/brpaz/lib-go/storage/db/log"
)

var (
	ErrDriverRequired     = errors.New("driver is required")
	ErrDsnRequired        = errors.New("dsn is required")
	ErrDriverNotSupported = errors.New("driver is not supported")
)

// ConnOpts contains the options for opening a new database connection.
type ConnOpts struct {
	Driver             string
	DSN                string
	OtelTracing        bool
	OtelMetrics        bool
	MaxIdleConns       int
	MaxOpenConns       int
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
	AutomaticPing      bool
	DefaultTransaction bool
	Logger             log.Logger
}

// Default connection options
var defaultConnOpts = ConnOpts{
	MaxIdleConns:       defaultMaxIdleConns,
	MaxOpenConns:       defaultMaxOpenConns,
	ConnMaxLifetime:    defaultConnMaxLifetime,
	ConnMaxIdleTime:    defaultConnMaxIdleTime,
	OtelTracing:        false,
	OtelMetrics:        false,
	AutomaticPing:      false,
	DefaultTransaction: false,
	Logger:             nil,
}

// Constants
const (
	DriverPostgres         = "postgres"
	defaultMaxOpenConns    = 10
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 1 * time.Hour
	defaultConnMaxIdleTime = 15 * time.Minute
)

var supportedDrivers = map[string]struct{}{
	DriverPostgres: {},
}

// Validate validates the connection options.
func (c *ConnOpts) Validate() error {
	if c.Driver == "" {
		return ErrDriverRequired
	}

	if c.DSN == "" {
		return ErrDsnRequired
	}

	if _, ok := supportedDrivers[c.Driver]; !ok {
		return ErrDriverNotSupported
	}

	return nil
}

// ConnOptFunc is a function that modifies the ConnOpts.
type ConnOptFunc func(*ConnOpts)

// WithDriver sets the database driver to use.
func WithDriver(driver string) ConnOptFunc {
	return func(o *ConnOpts) {
		o.Driver = driver
	}
}

// WithDSN sets the database DSN to use.
func WithDSN(dsn string) ConnOptFunc {
	return func(o *ConnOpts) {
		o.DSN = dsn
	}
}

// WithOtelTracing enables OpenTelemetry tracing for the database connection.
func WithOtelTracing() ConnOptFunc {
	return func(o *ConnOpts) {
		o.OtelTracing = true
	}
}

// WithOtelMetrics enables OpenTelemetry metrics for the database connection.
func WithOtelMetrics() ConnOptFunc {
	return func(o *ConnOpts) {
		o.OtelMetrics = true
	}
}

// WithMaxIdleConns sets the maximum number of idle connections in the connection pool.
func WithMaxIdleConns(maxIdleConns int) ConnOptFunc {
	return func(o *ConnOpts) {
		o.MaxIdleConns = maxIdleConns
	}
}

// WithMaxOpenConns sets the maximum number of open connections in the connection pool.
func WithMaxOpenConns(maxOpenConns int) ConnOptFunc {
	return func(o *ConnOpts) {
		o.MaxOpenConns = maxOpenConns
	}
}

// WithConnMaxLifetime sets the maximum amount of time a connection may be reused.
func WithConnMaxLifetime(connMaxLifetime time.Duration) ConnOptFunc {
	return func(o *ConnOpts) {
		o.ConnMaxLifetime = connMaxLifetime
	}
}

// WithConnMaxIdleTime sets the maximum amount of time a connection may be idle before being closed.
func WithConnMaxIdleTime(connMaxIdleTime time.Duration) ConnOptFunc {
	return func(o *ConnOpts) {
		o.ConnMaxIdleTime = connMaxIdleTime
	}
}

// WithAutomaticPing enables automatic pinging of the database connection.
func WithAutomaticPing() ConnOptFunc {
	return func(o *ConnOpts) {
		o.AutomaticPing = true
	}
}

// WithDefaultTransaction enables the default transaction mode for the connection.
func WithDefaultTransaction() ConnOptFunc {
	return func(o *ConnOpts) {
		o.DefaultTransaction = true
	}
}

// WithLogger sets the logger to be used by the connection.
func WithLogger(logger log.Logger) ConnOptFunc {
	return func(o *ConnOpts) {
		o.Logger = logger
	}
}

// NewConnection opens a new database connection using the provided options.
func NewConnection(opts ...ConnOptFunc) (*gorm.DB, error) {
	connOpts := defaultConnOpts

	// Apply options passed to the function
	for _, opt := range opts {
		opt(&connOpts)
	}

	if err := connOpts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid connection options: %w", err)
	}

	// Use a switch for future driver extensibility
	var dialector gorm.Dialector
	switch connOpts.Driver {
	case DriverPostgres:
		dialector = postgres.Open(connOpts.DSN)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", connOpts.Driver)
	}

	logger := gorml.Default.LogMode(gorml.Silent)
	if connOpts.Logger != nil {
		logger = dbLogger.NewGormLogger(connOpts.Logger)
	}

	// Open the connection using gorm
	gormDB, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: !connOpts.DefaultTransaction,
		DisableAutomaticPing:   !connOpts.AutomaticPing,
		Logger:                 logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if connOpts.OtelTracing {
		if err := setupOtelTracing(gormDB, connOpts); err != nil {
			return nil, err
		}
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying database connection: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(connOpts.MaxIdleConns)
	sqlDB.SetMaxOpenConns(connOpts.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(connOpts.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connOpts.ConnMaxIdleTime)

	return gormDB, nil
}

func setupOtelTracing(db *gorm.DB, opts ConnOpts) error {
	tracingOpts := []tracing.Option{}

	if !opts.OtelMetrics {
		tracingOpts = append(tracingOpts, tracing.WithoutMetrics())
	}

	if err := db.Use(tracing.NewPlugin(tracingOpts...)); err != nil {
		return fmt.Errorf("failed to setup tracing plugin: %w", err)
	}

	return nil
}
