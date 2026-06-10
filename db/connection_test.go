package db_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/db"
)

// sqliteDSN returns a fresh, isolated in-memory SQLite DSN.
func sqliteDSN(t *testing.T) string {
	t.Helper()
	return "file:" + t.Name() + "?mode=memory&cache=shared"
}

func TestNewConnection(t *testing.T) {
	t.Parallel()

	t.Run("success with valid dialector", func(t *testing.T) {
		t.Parallel()

		gdb, err := db.NewConnection(sqlite.Open(sqliteDSN(t)), db.WithMaxOpenConns(1))
		require.NoError(t, err)
		require.NotNil(t, gdb)

		sqlDB, err := gdb.DB()
		require.NoError(t, err)
		assert.NoError(t, sqlDB.Ping())
	})

	t.Run("success with log enabled", func(t *testing.T) {
		t.Parallel()

		gdb, err := db.NewConnection(
			sqlite.Open(sqliteDSN(t)),
			db.WithMaxOpenConns(1),
			db.WithQueryLog(),
		)
		require.NoError(t, err)
		require.NotNil(t, gdb)
	})

	t.Run("success with custom pool settings", func(t *testing.T) {
		t.Parallel()

		gdb, err := db.NewConnection(
			sqlite.Open(sqliteDSN(t)),
			db.WithMaxIdleConns(2),
			db.WithMaxOpenConns(5),
		)
		require.NoError(t, err)

		sqlDB, err := gdb.DB()
		require.NoError(t, err)
		stats := sqlDB.Stats()
		assert.LessOrEqual(t, stats.MaxOpenConnections, 5)
	})

	t.Run("fails with invalid pool settings", func(t *testing.T) {
		t.Parallel()

		_, err := db.NewConnection(
			sqlite.Open(sqliteDSN(t)),
			db.WithMaxIdleConns(-1),
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "MaxIdleConns must be >= 0")
	})

	t.Run("fails with unreachable host when ping enabled", func(t *testing.T) {
		t.Parallel()

		_, err := db.NewConnection(
			sqlite.Open("/nonexistent-dir/test.db"),
			db.WithAutomaticPing(),
		)
		require.Error(t, err)
	})
}
