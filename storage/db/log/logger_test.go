package log_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	gorml "gorm.io/gorm/logger"

	"github.com/brpaz/lib-go/log"
	dblog "github.com/brpaz/lib-go/storage/db/log"
	dbtestutil "github.com/brpaz/lib-go/storage/db/testutil"
)

func TestNewGormLogger(t *testing.T) {
	t.Parallel()

	t.Run("WithDefaultOpts", func(t *testing.T) {
		t.Parallel()
		logger := dblog.NewGormLogger(log.NewInMemory(log.LevelError))
		assert.IsType(t, logger, &dblog.GormLogger{})
		assert.Equal(t, gorml.Error, logger.LogLevel())
	})

	t.Run("WithCustomOpts", func(t *testing.T) {
		t.Parallel()
		logger := dblog.NewGormLogger(
			log.NewInMemory(log.LevelError),
			dblog.WithSilent(false),
			dblog.WithSlowThreshold(200*time.Millisecond),
			dblog.WithIgnoreRecordNotFoundError(false),
		)
		assert.IsType(t, logger, &dblog.GormLogger{})
		assert.Equal(t, gorml.Error, logger.LogLevel())
	})
}

func TestLogMode(t *testing.T) {
	t.Parallel()

	dbLogger := dblog.NewGormLogger(log.NewInMemory(log.LevelError))
	newL := dbLogger.LogMode(gorml.Error)

	assert.IsType(t, newL, &dblog.GormLogger{})

	logger := newL.(*dblog.GormLogger)
	assert.Equal(t, gorml.Error, logger.LogLevel())
}

func TestMapLogLevels(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		level    log.Level
		expected gorml.LogLevel
	}{
		{
			name:     "Debug",
			level:    log.LevelDebug,
			expected: gorml.Info,
		},
		{
			name:     "Info",
			level:    log.LevelInfo,
			expected: gorml.Info,
		},
		{
			name:     "Warn",
			level:    log.LevelWarn,
			expected: gorml.Warn,
		},
		{
			name:     "Error",
			level:    log.LevelError,
			expected: gorml.Error,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			logger := dblog.NewGormLogger(log.NewInMemory(tc.level))
			assert.Equal(t, tc.expected, logger.LogLevel())
		})
	}
}

func TestInfo(t *testing.T) {
	t.Parallel()

	t.Run("ShouldLog", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger)
		logger.Info(context.Background(), "Test")

		assert.Equal(t, 1, len(inMemoryLogger.Entries()))
		assert.Equal(t, "Test", inMemoryLogger.Entries()[0].Message)
	})

	t.Run("ShouldNotLog", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelError)
		logger := dblog.NewGormLogger(inMemoryLogger)
		logger.Info(context.Background(), "Test")

		assert.Equal(t, 0, len(inMemoryLogger.Entries()))
	})

	t.Run("ShouldNotLogSilent", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger, dblog.WithSilent(true))
		logger.Info(context.Background(), "Test")

		assert.Equal(t, 0, len(inMemoryLogger.Entries()))
	})
}

func TestWarn(t *testing.T) {
	t.Parallel()

	t.Run("ShouldLog", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger)
		logger.Warn(context.Background(), "Test")

		assert.Equal(t, 1, len(inMemoryLogger.Entries()))
		assert.Equal(t, "Test", inMemoryLogger.Entries()[0].Message)
	})

	t.Run("ShouldNotLog", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelError)
		logger := dblog.NewGormLogger(inMemoryLogger)
		logger.Warn(context.Background(), "Test")

		assert.Equal(t, 0, len(inMemoryLogger.Entries()))
	})

	t.Run("ShouldNotLogSilent", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger, dblog.WithSilent(true))
		logger.Warn(context.Background(), "Test")

		assert.Equal(t, 0, len(inMemoryLogger.Entries()))
	})
}

func TestError(t *testing.T) {
	t.Parallel()

	t.Run("ShouldLog", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger)
		logger.Error(context.Background(), "Test")

		assert.Equal(t, 1, len(inMemoryLogger.Entries()))
		assert.Equal(t, "Test", inMemoryLogger.Entries()[0].Message)
	})

	t.Run("ShouldNotLogSilent", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelError)
		logger := dblog.NewGormLogger(inMemoryLogger, dblog.WithSilent(true))
		logger.Error(context.Background(), "Test")

		assert.Equal(t, 0, len(inMemoryLogger.Entries()))
	})
}

func TestTrace(t *testing.T) {
	t.Parallel()
	t.Run("ShouldLogError", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger)
		logger.Trace(context.Background(), time.Now(), func() (string, int64) {
			return "SELECT * FROM users", 0
		}, assert.AnError)

		assert.Equal(t, 1, len(inMemoryLogger.Entries()))
		assert.Equal(t, "SQL execution error", inMemoryLogger.Entries()[0].Message)
	})

	t.Run("ShouldLogSlowQuery", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger, dblog.WithSlowThreshold(1*time.Millisecond))
		logger.Trace(context.Background(), time.Now().Add(-1*time.Second), func() (string, int64) {
			return "SELECT * FROM users", 0
		}, nil)

		assert.Equal(t, 1, len(inMemoryLogger.Entries()))
		assert.Equal(t, "Slow SQL query", inMemoryLogger.Entries()[0].Message)
	})

	t.Run("ShouldLogTrace", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger)
		logger.Trace(context.Background(), time.Now(), func() (string, int64) {
			return "SELECT * FROM users", 0
		}, nil)

		assert.Equal(t, 1, len(inMemoryLogger.Entries()))
		assert.Equal(t, "SQL query trace", inMemoryLogger.Entries()[0].Message)
	})

	t.Run("ShouldNotLogSilent", func(t *testing.T) {
		t.Parallel()
		inMemoryLogger := log.NewInMemory(log.LevelDebug)
		logger := dblog.NewGormLogger(inMemoryLogger, dblog.WithSilent(true))
		logger.Trace(context.Background(), time.Now(), func() (string, int64) {
			return "SELECT * FROM users", 0
		}, nil)

		assert.Equal(t, 0, len(inMemoryLogger.Entries()))
	})
}

func TestTrace_Integration(t *testing.T) {
	ctx := context.Background()
	dbContainer, iniErr := dbtestutil.InitPgTestContainer(ctx)
	require.NoError(t, iniErr)

	defer func() {
		_ = dbContainer.Stop(ctx)
	}()

	dsn, dsnErr := dbContainer.GetDSN(ctx)
	require.NoError(t, dsnErr)

	t.Run("Log Query", func(t *testing.T) {
		memoryLogger := log.NewInMemory(log.LevelDebug)

		gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: dblog.NewGormLogger(memoryLogger),
		})
		require.NoError(t, err)

		_, err = gormDB.Raw("SELECT 1").Rows()
		require.NoError(t, err)

		assert.Equal(t, 1, len(memoryLogger.Entries()))

		entry := memoryLogger.Entries()[0]
		assert.Equal(t, "SQL query trace", entry.Message)

		queryField, ok := entry.GetField("sql")
		require.True(t, ok)
		assert.Equal(t, "SELECT 1", queryField.String)

		_, ok = entry.GetField("elapsed")
		require.True(t, ok)
	})

	t.Run("Log Slow Query", func(t *testing.T) {
		memoryLogger := log.NewInMemory(log.LevelDebug)

		gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: dblog.NewGormLogger(
				memoryLogger,
				dblog.WithSlowThreshold(1*time.Millisecond),
			),
		})
		require.NoError(t, err)

		_, err = gormDB.Raw("SELECT pg_sleep(1)").Rows()
		require.NoError(t, err)

		assert.Equal(t, 1, len(memoryLogger.Entries()))

		entry := memoryLogger.Entries()[0]
		assert.Equal(t, "Slow SQL query", entry.Message)

		queryField, ok := entry.GetField("sql")
		require.True(t, ok)
		assert.Equal(t, "SELECT pg_sleep(1)", queryField.String)

		_, ok = entry.GetField("elapsed")
		require.True(t, ok)
	})
}
