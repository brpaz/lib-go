package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/log"
	"github.com/brpaz/lib-go/storage/db"
)

func TestConnOpts_Validate(t *testing.T) {
	t.Parallel()

	defaultOpts := db.ConnOpts{
		Driver: "postgres",
		DSN:    "postgres://user:password@localhost:5432/db",
	}

	t.Run("ValidOpts", func(t *testing.T) {
		t.Parallel()
		opts := defaultOpts
		err := opts.Validate()

		assert.Nil(t, err)
	})

	t.Run("MissingDriver", func(t *testing.T) {
		t.Parallel()
		opts := defaultOpts
		opts.Driver = ""

		err := opts.Validate()

		assert.NotNil(t, err)
		assert.Equal(t, db.ErrDriverRequired, err)
	})

	t.Run("MissingDSN", func(t *testing.T) {
		t.Parallel()
		opts := defaultOpts
		opts.DSN = ""

		err := opts.Validate()

		assert.NotNil(t, err)
		assert.Equal(t, db.ErrDsnRequired, err)
	})

	t.Run("UnsupportedDriver", func(t *testing.T) {
		t.Parallel()
		opts := defaultOpts
		opts.Driver = "invalid"

		err := opts.Validate()

		assert.NotNil(t, err)
		assert.Equal(t, db.ErrDriverNotSupported, err)
	})
}

func TestNewConnection(t *testing.T) {
	t.Parallel()

	dsn, err := dbInstance.GetDSN(context.Background())
	require.NoError(t, err)

	t.Run("WithInvalidOpts", func(t *testing.T) {
		t.Parallel()
		_, err := db.NewConnection()
		require.Error(t, err)
	})

	t.Run("WithDefaultOpts", func(t *testing.T) {
		t.Parallel()
		gormConn, err := db.NewConnection(
			db.WithDSN(dsn),
			db.WithDriver(db.DriverPostgres),
		)
		require.NoError(t, err)
		assert.NotNil(t, gormConn)

		sqlDB, err := gormConn.DB()
		require.NoError(t, err)

		_, err = sqlDB.Query("SELECT 1")
		assert.NoError(t, err)

		assert.True(t, gormConn.Config.SkipDefaultTransaction)
		assert.True(t, gormConn.Config.DisableAutomaticPing)
		assert.Equal(t, 10, sqlDB.Stats().MaxOpenConnections)
	})

	t.Run("WithCustomOpts", func(t *testing.T) {
		t.Parallel()
		gormConn, err := db.NewConnection(
			db.WithLogger(log.NewInMemory(log.LevelDebug)),
			db.WithDSN(dsn),
			db.WithDriver(db.DriverPostgres),
			db.WithMaxIdleConns(1),
			db.WithMaxOpenConns(1),
			db.WithConnMaxIdleTime(1*time.Minute),
			db.WithConnMaxLifetime(5*time.Minute),
			db.WithAutomaticPing(),
			db.WithDefaultTransaction(),
			db.WithOtelTracing(),
			db.WithOtelMetrics(),
		)
		require.NoError(t, err)
		assert.NotNil(t, gormConn)

		sqlDB, err := gormConn.DB()
		require.NoError(t, err)

		_, err = sqlDB.Query("SELECT 1")
		assert.NoError(t, err)

		assert.False(t, gormConn.Config.SkipDefaultTransaction)
		assert.False(t, gormConn.Config.DisableAutomaticPing)
		assert.Equal(t, 1, sqlDB.Stats().MaxOpenConnections)
	})

	t.Run("WithInvalidCredentials", func(t *testing.T) {
		t.Parallel()
		_, err := db.NewConnection(
			db.WithAutomaticPing(),
			db.WithDSN("postgres://invalid:invalid@localhost:5432/db"),
			db.WithDriver(db.DriverPostgres),
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open database connection")
	})
}
