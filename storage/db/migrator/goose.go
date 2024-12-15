package migrator

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/pressly/goose/v3"
)

var (
	ErrDBNotSet              = errors.New("a sql.DB connection is required")
	ErrMigrationsNotSet      = errors.New("migrations directory or filesystem is required")
	ErrInvalidMigrationsType = errors.New("invalid migrations type. must be either 'go' or 'sql'")
)

const (
	MigrationTypeGo  = "go"
	MigrationTypeSQL = "sql"
)

// GooseMigrator is a migrator that uses Goose under the hood.
type GooseMigrator struct {
	// DB is the database connection to use for migrations.
	DB *sql.DB

	// Dir is the directory containing migration files.
	MigrationsDir string

	// FS is the filesystem containing migration files.
	MigrationsFS embed.FS

	// MigrationsType is the type of migrations to run.
	MigrationsType string

	// Dialect is the database dialect to use for migrations.
	Dialect string

	// AllowOutOfOrder allows migrations
	AllowOutOfOrder bool

	// Sequntial defines to use sequential migrations instead of timestamped
	Sequencial bool
}

func (m *GooseMigrator) Validate() error {
	if m.DB == nil {
		return ErrDBNotSet
	}

	if m.MigrationsDir == "" {
		return ErrMigrationsNotSet
	}

	if m.MigrationsType != MigrationTypeGo && m.MigrationsType != MigrationTypeSQL {
		return ErrInvalidMigrationsType
	}

	return nil
}

// GooseMigratorOpt is a functional option for configuring a GooseMigrator.
type GooseMigratorOpt func(*GooseMigrator)

// WithGooseMigrationsDir sets the directory containing migration files.
func WithGooseMigrationsDir(dir string) GooseMigratorOpt {
	return func(m *GooseMigrator) {
		m.MigrationsDir = dir
	}
}

// WithGooseMigrationsFS sets the filesystem containing migration files.
func WithGooseMigrationsFS(fs embed.FS) GooseMigratorOpt {
	return func(m *GooseMigrator) {
		m.MigrationsFS = fs
		goose.SetBaseFS(m.MigrationsFS)
	}
}

// WithGooseMigrationsType sets the type of migrations to run.
func WithGooseMigrationsType(t string) GooseMigratorOpt {
	return func(m *GooseMigrator) {
		m.MigrationsType = t
	}
}

// WithGooseDB sets the database connection to use for migrations.
func WithGooseDB(db *sql.DB) GooseMigratorOpt {
	return func(m *GooseMigrator) {
		m.DB = db
	}
}

// WithGooseDialect sets the database dialect to use for migrations.
func WithGooseDialect(dialect string) GooseMigratorOpt {
	return func(m *GooseMigrator) {
		m.Dialect = dialect
		_ = goose.SetDialect(m.Dialect)
	}
}

// WithGooseAllowOutOfOrder allows migrations to run out of order.
func WithGooseAllowOutOfOrder(allow bool) GooseMigratorOpt {
	return func(m *GooseMigrator) {
		m.AllowOutOfOrder = allow
	}
}

// WithGooseSequencial allows to use sequential migrations instead of timestamped
func WithGooseSequencial(sequencial bool) GooseMigratorOpt {
	return func(m *GooseMigrator) {
		m.Sequencial = sequencial
		goose.SetSequential(sequencial)
	}
}

// NewGooseMigrator creates a new GooseMigrator with the given options.
func NewGooseMigrator(opts ...GooseMigratorOpt) (*GooseMigrator, error) {
	migrator := &GooseMigrator{
		MigrationsType:  MigrationTypeSQL,
		AllowOutOfOrder: false,
	}

	for _, opt := range opts {
		opt(migrator)
	}

	if err := migrator.Validate(); err != nil {
		return nil, err
	}

	return migrator, nil
}

// Up runs all available migrations.
func (m *GooseMigrator) Up(ctx context.Context) error {
	gooseOpts := []goose.OptionsFunc{}

	if m.AllowOutOfOrder {
		gooseOpts = append(gooseOpts, goose.WithAllowMissing())
	}

	err := goose.UpContext(ctx, m.DB, m.MigrationsDir, gooseOpts...)

	return err
}

// Down rolls back the most recent migration.
func (m *GooseMigrator) Down(ctx context.Context) error {
	err := goose.DownContext(ctx, m.DB, m.MigrationsDir)
	return err
}

// Reset rolls back all migrations.
func (m *GooseMigrator) Reset(ctx context.Context) error {
	err := goose.ResetContext(ctx, m.DB, m.MigrationsDir)
	return err
}

// Create creates a new migration file.
func (m *GooseMigrator) Create(_ context.Context, name string) error {
	err := goose.Create(m.DB, m.MigrationsDir, name, m.MigrationsType)
	return err
}
