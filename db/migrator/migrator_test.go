package migrator_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/db/dbtest"
	"github.com/brpaz/lib-go/db/migrator"
)

// testMigrations is a minimal set of goose migrations used to exercise the
// Migrator independently of any application's schema.
var testMigrations = os.DirFS("testdata/migrations")

// testPG is a package-scoped Postgres container shared across all tests.
var testPG *dbtest.PostgresContainer

func TestMain(m *testing.M) {
	var err error
	testPG, err = dbtest.NewPostgresContainer(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	testPG.Terminate(context.Background())
	os.Exit(code)
}

// newPostgresDB returns a fresh, isolated Postgres database for use in tests.
func newPostgresDB(t *testing.T) *sql.DB {
	t.Helper()
	return testPG.NewIsolatedDB(t)
}

// newSQLiteDB returns a fresh, isolated in-memory SQLite database for use in tests.
func newSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()

	gdb, err := dbtest.NewSQLite()
	require.NoError(t, err, "open sqlite db")

	sqlDB, err := gdb.DB()
	require.NoError(t, err, "retrieve sql.DB")

	t.Cleanup(func() { _ = sqlDB.Close() })

	return sqlDB
}

// TestMigrator_SQLite covers dialect-agnostic Migrator behavior against an
// in-memory SQLite database, which requires no Docker.
func TestMigrator_SQLite(t *testing.T) {
	// Do NOT add t.Parallel() here or to subtests: goose uses package-level
	// global state (SetBaseFS, SetDialect). Parallel subtests would race on
	// that shared state.

	t.Run("runs pending migrations", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect(migrator.DialectSqlite))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
	})

	t.Run("migrate is idempotent", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect(migrator.DialectSqlite))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
		assert.NoError(t, m.Migrate(context.Background()))
	})

	t.Run("rollback reverts last migration", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect(migrator.DialectSqlite))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
		require.NoError(t, m.Rollback(context.Background()))
	})

	t.Run("status reports applied migrations", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect(migrator.DialectSqlite))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
		assert.NoError(t, m.Status(context.Background()))
	})

	t.Run("fails with invalid dialect", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		_, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect("notadialect"))
		require.Error(t, err)
		assert.ErrorContains(t, err, "set dialect")
	})
}

// TestMigrator_Postgres covers Postgres-specific wiring (real DSN, dialect
// registration) against a Testcontainers Postgres instance.
func TestMigrator_Postgres(t *testing.T) {
	// Do NOT add t.Parallel() here or to subtests: goose uses package-level
	// global state (SetBaseFS, SetDialect). Parallel subtests would race on
	// that shared state.

	t.Run("runs pending migrations", func(t *testing.T) {
		sqlDB := newPostgresDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect(migrator.DialectPostgres))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
	})

	t.Run("status reports applied migrations", func(t *testing.T) {
		sqlDB := newPostgresDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect(migrator.DialectPostgres))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
		assert.NoError(t, m.Status(context.Background()))
	})
}
