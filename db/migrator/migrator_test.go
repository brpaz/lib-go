package migrator_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"

	"github.com/brpaz/lib-go/db/migrator"
)

// testMigrations is a minimal set of goose migrations used to exercise the
// Migrator independently of any application's schema.
var testMigrations = os.DirFS("testdata/migrations")

// newSQLiteDB opens a fresh, isolated in-memory SQLite database.
func newSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()

	sqlDB, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	require.NoError(t, err, "open sqlite db")
	sqlDB.SetMaxOpenConns(1)

	t.Cleanup(func() { _ = sqlDB.Close() })

	return sqlDB
}

func TestMigrator(t *testing.T) {
	// Do NOT add t.Parallel() here or to subtests: goose uses package-level
	// global state (SetBaseFS, SetDialect). Parallel subtests would race on
	// that shared state.

	t.Run("runs pending migrations", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect("sqlite3"))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
	})

	t.Run("migrate is idempotent", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect("sqlite3"))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
		assert.NoError(t, m.Migrate(context.Background()))
	})

	t.Run("rollback reverts last migration", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect("sqlite3"))
		require.NoError(t, err)

		require.NoError(t, m.Migrate(context.Background()))
		require.NoError(t, m.Rollback(context.Background()))
	})

	t.Run("status reports applied migrations", func(t *testing.T) {
		sqlDB := newSQLiteDB(t)
		m, err := migrator.NewMigrator(sqlDB, testMigrations, migrator.WithDialect("sqlite3"))
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
