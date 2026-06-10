package db_test

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/brpaz/lib-go/db"
)

// connSeq ensures every sqliteDialector call gets its own isolated in-memory
// database, even when multiple tests run in parallel.
var connSeq atomic.Uint64

// sqliteDialector returns a dialector for a unique in-memory SQLite database
// for use in a single test.
func sqliteDialector(t *testing.T) gorm.Dialector {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", name, connSeq.Add(1))
	return sqlite.Open(dsn)
}

func TestNewConnection(t *testing.T) {
	t.Parallel()

	t.Run("success with valid dialector", func(t *testing.T) {
		t.Parallel()

		gdb, err := db.NewConnection(sqliteDialector(t), db.WithMaxOpenConns(1))
		require.NoError(t, err)
		require.NotNil(t, gdb)

		sqlDB, err := gdb.DB()
		require.NoError(t, err)
		assert.NoError(t, sqlDB.Ping())
	})

	t.Run("success with log enabled", func(t *testing.T) {
		t.Parallel()

		gdb, err := db.NewConnection(
			sqliteDialector(t),
			db.WithMaxOpenConns(1),
			db.WithQueryLog(),
		)
		require.NoError(t, err)
		require.NotNil(t, gdb)
	})

	t.Run("success with custom pool settings", func(t *testing.T) {
		t.Parallel()

		gdb, err := db.NewConnection(
			sqliteDialector(t),
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
			sqliteDialector(t),
			db.WithMaxIdleConns(-1),
		)
		require.Error(t, err)
		assert.ErrorContains(t, err, "MaxIdleConns must be >= 0")
	})

	t.Run("fails with unreachable host when ping enabled", func(t *testing.T) {
		t.Parallel()

		_, err := db.NewConnection(
			sqlite.Open("file:/nonexistent/dir/db.sqlite?mode=ro"),
			db.WithAutomaticPing(),
		)
		require.Error(t, err)
	})
}
