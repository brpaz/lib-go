package dbtest

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/brpaz/lib-go/db/migrator"
)

const (
	defaultPostgresImage    = "postgres:18-alpine"
	defaultPostgresDatabase = "testdb"
	defaultPostgresUser     = "postgres"
	defaultPostgresPassword = "postgres"
)

// PostgresOptions configures [NewPostgresContainer].
type PostgresOptions struct {
	Image         string
	Database      string
	Username      string
	Password      string
	ContainerOpts []testcontainers.ContainerCustomizer
}

// PostgresOption is a functional option for [NewPostgresContainer].
type PostgresOption func(*PostgresOptions)

// WithPostgresImage overrides the Postgres image used by the container.
// Defaults to "postgres:18-alpine".
func WithPostgresImage(image string) PostgresOption {
	return func(o *PostgresOptions) { o.Image = image }
}

// WithPostgresDatabase overrides the default database name. Defaults to "testdb".
func WithPostgresDatabase(name string) PostgresOption {
	return func(o *PostgresOptions) { o.Database = name }
}

// WithPostgresUsername overrides the default Postgres username. Defaults to "postgres".
func WithPostgresUsername(name string) PostgresOption {
	return func(o *PostgresOptions) { o.Username = name }
}

// WithPostgresPassword overrides the default Postgres password. Defaults to "postgres".
func WithPostgresPassword(password string) PostgresOption {
	return func(o *PostgresOptions) { o.Password = password }
}

// WithPostgresContainerOptions appends additional Testcontainers customizers
// (e.g. custom wait strategies, env vars, mounts) to the container request.
func WithPostgresContainerOptions(opts ...testcontainers.ContainerCustomizer) PostgresOption {
	return func(o *PostgresOptions) { o.ContainerOpts = append(o.ContainerOpts, opts...) }
}

// PostgresContainer wraps a Testcontainers PostgreSQL container for use in tests.
type PostgresContainer struct {
	DSN string
	ctr *tcpostgres.PostgresContainer
}

// NewPostgresContainer starts a PostgreSQL container and returns it.
// The caller is responsible for terminating it — call [PostgresContainer.Cleanup] to
// register automatic termination scoped to a test, or [PostgresContainer.Terminate]
// after m.Run() in TestMain.
func NewPostgresContainer(ctx context.Context, opts ...PostgresOption) (c *PostgresContainer, err error) {
	o := &PostgresOptions{
		Image:    defaultPostgresImage,
		Database: defaultPostgresDatabase,
		Username: defaultPostgresUser,
		Password: defaultPostgresPassword,
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
		return nil, fmt.Errorf("dbtest: failed to start postgres container: %w", err)
	}
	defer func() {
		if err != nil {
			_ = testcontainers.TerminateContainer(ctr)
		}
	}()

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("dbtest: failed to get connection string: %w", err)
	}

	c = &PostgresContainer{DSN: dsn, ctr: ctr}

	return c, nil
}

// MigratePostgres opens a raw *sql.DB connection to dsn and applies the
// goose SQL migrations from fsys using the postgres dialect.
//
// Use this against a [PostgresContainer.DSN] (or a database returned by
// [PostgresContainer.NewIsolatedDB]) to prepare a schema before running tests.
func MigratePostgres(ctx context.Context, dsn string, fsys fs.FS) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("dbtest: open db for migrations: %w", err)
	}
	defer sqlDB.Close()

	m, err := migrator.NewMigrator(sqlDB, fsys, migrator.WithDialect(migrator.DialectPostgres))
	if err != nil {
		return fmt.Errorf("dbtest: create migrator: %w", err)
	}

	if err := m.Migrate(ctx); err != nil {
		return fmt.Errorf("dbtest: run migrations: %w", err)
	}

	return nil
}

// Cleanup registers a t.Cleanup hook that terminates the container when t ends.
// Call this immediately after [NewPostgresContainer] in per-test usage.
func (c *PostgresContainer) Cleanup(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { c.Terminate(context.Background()) })
}

// Terminate stops the container. Call this after m.Run() returns in TestMain.
func (c *PostgresContainer) Terminate(ctx context.Context) {
	if err := testcontainers.TerminateContainer(c.ctr); err != nil {
		log.Printf("dbtest: failed to terminate postgres container: %v", err)
	}
}

// OpenDB opens a *sql.DB connection to the container using the pgx driver.
// The connection is closed automatically when t ends via t.Cleanup.
func (c *PostgresContainer) OpenDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", c.DSN)
	require.NoError(t, err, "dbtest: failed to open sql.DB")

	t.Cleanup(func() { _ = db.Close() })

	return db
}

// NewIsolatedDB creates a fresh Postgres database on the shared container and
// returns an open *sql.DB connected to it. The database is dropped when t ends.
//
// Use this in package-scoped container tests to keep each test's schema changes
// independent without spinning up a new container per test.
func (c *PostgresContainer) NewIsolatedDB(t *testing.T) *sql.DB {
	t.Helper()

	dbName := isolatedDBName(t.Name())

	// Use the container's default database to create the isolated one.
	adminDB, err := sql.Open("pgx", c.DSN)
	require.NoError(t, err, "dbtest: failed to open admin connection")
	defer adminDB.Close()

	_, err = adminDB.Exec(fmt.Sprintf(`CREATE DATABASE %q`, dbName))
	require.NoError(t, err, "dbtest: failed to create isolated database %q", dbName)

	// Build a DSN pointing to the new database.
	u, err := url.Parse(c.DSN)
	require.NoError(t, err, "dbtest: failed to parse DSN")
	u.Path = "/" + dbName
	isolatedDSN := u.String()

	db, err := sql.Open("pgx", isolatedDSN)
	require.NoError(t, err, "dbtest: failed to open isolated sql.DB")

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
