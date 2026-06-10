package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

const (
	defaultTableName = "schema_migrations"
	DialectPostgres  = "postgres"
	DialectSqlite    = "sqlite3"
)

// Option is a functional option for configuring the Migrator.
type Option func(*Migrator)

// WithMigrationsDir sets the directory within fsys that contains the SQL files.
// Defaults to "." (root of fsys). Use this when the embed FS contains a named
// sub-directory, e.g. WithMigrationsDir("migrations").
func WithMigrationsDir(dir string) Option {
	return func(m *Migrator) {
		m.migrationsDir = dir
	}
}

func WithSchemaTableName(name string) Option {
	return func(m *Migrator) {
		m.migrationsTableName = name
	}
}

// WithDialect sets the SQL dialect used by goose. Defaults to "postgres".
// Valid values are the dialect names accepted by goose.SetDialect.
func WithDialect(dialect string) Option {
	return func(m *Migrator) {
		m.dialect = dialect
	}
}

// WithLogger sets the logger used for migration output.
// Use [NewLibLogger] or [NewNopLogger], or supply any [Logger] implementation.
func WithLogger(logger Logger) Option {
	return func(m *Migrator) {
		m.logger = logger
	}
}

// Migrator runs database schema migrations using goose.
//
// Note: goose uses package-level global state (base FS, dialect, table name).
// Avoid running multiple Migrator instances concurrently with different options.
type Migrator struct {
	db                  *sql.DB
	fsys                fs.FS
	migrationsDir       string
	migrationsTableName string
	dialect             string
	logger              Logger
}

// NewMigrator constructs a Migrator. fsys must contain the SQL migration files.
// By default migrations are read from the root of fsys; use [WithMigrationsDir]
// to specify a sub-directory (e.g. when using //go:embed migrations/*.sql from
// a parent package).
func NewMigrator(db *sql.DB, fsys fs.FS, opts ...Option) (*Migrator, error) {
	m := &Migrator{
		db:                  db,
		fsys:                fsys,
		migrationsDir:       ".",
		migrationsTableName: defaultTableName,
		dialect:             DialectPostgres,
		logger:              NewNopLogger(),
	}
	for _, opt := range opts {
		opt(m)
	}

	if err := m.setup(); err != nil {
		return nil, err
	}

	return m, nil
}

// Migrate runs all pending UP migrations.
func (m *Migrator) Migrate(ctx context.Context) error {
	if err := goose.UpContext(ctx, m.db, m.migrationsDir); err != nil {
		return fmt.Errorf("migrator: %w", err)
	}
	return nil
}

// Rollback rolls back the last applied migration.
func (m *Migrator) Rollback(ctx context.Context) error {
	if err := goose.DownContext(ctx, m.db, m.migrationsDir); err != nil {
		return fmt.Errorf("migrator: %w", err)
	}
	return nil
}

// Status logs the current applied/pending state of all migrations.
func (m *Migrator) Status(ctx context.Context) error {
	if err := goose.StatusContext(ctx, m.db, m.migrationsDir); err != nil {
		return fmt.Errorf("migrator: %w", err)
	}
	return nil
}

// setup configures the goose package-level globals. Called once from NewMigrator.
func (m *Migrator) setup() error {
	goose.SetBaseFS(m.fsys)
	goose.SetTableName("schema_migrations")
	goose.SetLogger(m.logger)

	if err := goose.SetDialect(m.dialect); err != nil {
		return fmt.Errorf("migrator: set dialect %q: %w", m.dialect, err)
	}
	return nil
}
