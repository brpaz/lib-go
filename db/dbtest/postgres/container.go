package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/brpaz/lib-go/db"
	"github.com/brpaz/lib-go/db/migrator"
)

const (
	defaultImage    = "postgres:18-alpine"
	defaultDatabase = "testdb"
	defaultUser     = "postgres"
	defaultPassword = "postgres"
)

// Options configures the test container.
type Options struct {
	Image         string
	Database      string
	Username      string
	Password      string
	AutoMigrate   bool
	MigrationsFS  fs.FS
	MigrationsDir string
	ContainerOpts []testcontainers.ContainerCustomizer
}

// Option is a functional option for [NewTestContainer].
type Option func(*Options)

// WithImage overrides the Postgres image used by the container.
// Defaults to "postgres:18-alpine".
func WithImage(image string) Option {
	return func(o *Options) { o.Image = image }
}

// WithDatabase overrides the default database name. Defaults to "testdb".
func WithDatabase(name string) Option {
	return func(o *Options) { o.Database = name }
}

// WithUsername overrides the default Postgres username. Defaults to "postgres".
func WithUsername(name string) Option {
	return func(o *Options) { o.Username = name }
}

// WithPassword overrides the default Postgres password. Defaults to "postgres".
func WithPassword(password string) Option {
	return func(o *Options) { o.Password = password }
}

// WithAutoMigrate enables running all pending migrations against the
// container's default database after it starts. Requires [WithMigrationsFS]
// to also be set.
func WithAutoMigrate() Option {
	return func(o *Options) { o.AutoMigrate = true }
}

// WithMigrationsFS sets the filesystem containing goose SQL migration files.
// Has no effect unless [WithAutoMigrate] is also set.
func WithMigrationsFS(fsys fs.FS) Option {
	return func(o *Options) { o.MigrationsFS = fsys }
}

// WithMigrationsDir sets the directory within the migrations FS that contains
// the SQL files. Defaults to "." (root of fsys). Use this when the FS passed
// to [WithMigrationsFS] contains a named sub-directory, e.g. "migrations".
func WithMigrationsDir(dir string) Option {
	return func(o *Options) { o.MigrationsDir = dir }
}

// WithContainerOptions appends additional Testcontainers customizers
// (e.g. custom wait strategies, env vars, mounts) to the container request.
func WithContainerOptions(opts ...testcontainers.ContainerCustomizer) Option {
	return func(o *Options) { o.ContainerOpts = append(o.ContainerOpts, opts...) }
}

// Container wraps a Testcontainers PostgreSQL container for use in tests.
type Container struct {
	DSN string
	ctr *tcpostgres.PostgresContainer
}

// NewTestContainer starts a PostgreSQL container and returns it.
// The caller is responsible for terminating it — call [Container.Cleanup] to
// register automatic termination scoped to a test, or [Container.Terminate]
// after m.Run() in TestMain.
func NewTestContainer(ctx context.Context, opts ...Option) (c *Container, err error) {
	o := &Options{
		Image:    defaultImage,
		Database: defaultDatabase,
		Username: defaultUser,
		Password: defaultPassword,
	}
	for _, opt := range opts {
		opt(o)
	}

	runOpts := append([]testcontainers.ContainerCustomizer{
		tcpostgres.WithDatabase(o.Database),
		tcpostgres.WithUsername(o.Username),
		tcpostgres.WithPassword(o.Password),
		tcpostgres.BasicWaitStrategies(),
	}, o.ContainerOpts...)

	ctr, err := tcpostgres.Run(ctx, o.Image, runOpts...)
	if err != nil {
		return nil, fmt.Errorf("dbtest/postgres: failed to start postgres container: %w", err)
	}
	defer func() {
		if err != nil {
			_ = testcontainers.TerminateContainer(ctr)
		}
	}()

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("dbtest/postgres: failed to get connection string: %w", err)
	}

	c = &Container{DSN: dsn, ctr: ctr}

	if o.AutoMigrate {
		if o.MigrationsFS == nil {
			return nil, errors.New("dbtest/postgres: WithAutoMigrate requires WithMigrationsFS")
		}
		if err = c.Migrate(ctx, o); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Migrate opens a raw sql.DB, runs all pending migrations from o.MigrationsFS, then closes it.
func (c *Container) Migrate(ctx context.Context, o *Options) error {
	sqlDB, err := sql.Open("pgx", c.DSN)
	if err != nil {
		return fmt.Errorf("dbtest/postgres: open db for migrations: %w", err)
	}
	defer sqlDB.Close()

	var migratorOpts []migrator.Option
	if o.MigrationsDir != "" {
		migratorOpts = append(migratorOpts, migrator.WithMigrationsDir(o.MigrationsDir))
	}

	m, err := migrator.NewMigrator(sqlDB, o.MigrationsFS, migratorOpts...)
	if err != nil {
		return fmt.Errorf("dbtest/postgres: create migrator: %w", err)
	}

	if err := m.Migrate(ctx); err != nil {
		return fmt.Errorf("dbtest/postgres: run migrations: %w", err)
	}

	return nil
}

// GetGormConnection opens a GORM connection to the container.
func (c *Container) GetGormConnection() (*gorm.DB, error) {
	gormDB, err := db.NewConnection(gormpostgres.Open(c.DSN))
	if err != nil {
		return nil, fmt.Errorf("dbtest/postgres: open gorm connection: %w", err)
	}

	return gormDB, nil
}

// Cleanup registers a t.Cleanup hook that terminates the container when t ends.
// Call this immediately after [NewTestContainer] in per-test usage.
func (c *Container) Cleanup(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { c.Terminate(context.Background()) })
}

// Terminate stops the container. Call this after m.Run() returns in TestMain.
func (c *Container) Terminate(ctx context.Context) {
	if err := testcontainers.TerminateContainer(c.ctr); err != nil {
		log.Printf("dbtest/postgres: failed to terminate container: %v", err)
	}
}

// OpenDB opens a *sql.DB connection to the container using the pgx driver.
// The connection is closed automatically when t ends via t.Cleanup.
func (c *Container) OpenDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", c.DSN)
	require.NoError(t, err, "dbtest/postgres: failed to open sql.DB")

	t.Cleanup(func() { _ = db.Close() })

	return db
}

// NewIsolatedDB creates a fresh Postgres database on the shared container and
// returns an open *sql.DB connected to it. The database is dropped when t ends.
//
// Use this in package-scoped container tests to keep each test's schema changes
// independent without spinning up a new container per test.
func (c *Container) NewIsolatedDB(t *testing.T) *sql.DB {
	t.Helper()

	dbName := isolatedDBName(t.Name())

	// Use the container's default database to create the isolated one.
	adminDB, err := sql.Open("pgx", c.DSN)
	require.NoError(t, err, "dbtest/postgres: failed to open admin connection")
	defer adminDB.Close()

	_, err = adminDB.Exec(fmt.Sprintf(`CREATE DATABASE %q`, dbName))
	require.NoError(t, err, "dbtest/postgres: failed to create isolated database %q", dbName)

	// Build a DSN pointing to the new database.
	u, err := url.Parse(c.DSN)
	require.NoError(t, err, "dbtest/postgres: failed to parse DSN")
	u.Path = "/" + dbName
	isolatedDSN := u.String()

	db, err := sql.Open("pgx", isolatedDSN)
	require.NoError(t, err, "dbtest/postgres: failed to open isolated sql.DB")

	t.Cleanup(func() {
		_ = db.Close()
		// Re-open admin connection for the DROP (can't drop the current DB).
		adb, _ := sql.Open("pgx", c.DSN)
		defer adb.Close()
		_, _ = adb.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q WITH (FORCE)`, dbName))
	})

	return db
}

// isolatedDBName returns a valid, deterministic Postgres identifier derived from
// the test name. It lowercases, replaces non-alphanumeric characters with
// underscores, prefixes with "t_", and truncates to 63 characters.
func isolatedDBName(testName string) string {
	var b strings.Builder
	b.WriteString("t_")
	for _, r := range strings.ToLower(testName) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	s := b.String()
	if len(s) > 63 {
		s = s[:63]
	}
	return s
}
