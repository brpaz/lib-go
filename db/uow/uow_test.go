package uow_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/brpaz/lib-go/db/dbtest"
	"github.com/brpaz/lib-go/db/uow"
)

// repos is a minimal repository bundle used for testing.
type repos struct {
	DB *gorm.DB
}

func newGormDB(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := dbtest.NewSQLite()
	require.NoError(t, err)
	return gormDB
}

func TestNewManager(t *testing.T) {
	t.Parallel()

	t.Run("nil db returns error", func(t *testing.T) {
		t.Parallel()

		_, err := uow.NewManager(nil, func(tx *gorm.DB) repos { return repos{DB: tx} })
		require.Error(t, err)
		assert.ErrorContains(t, err, "db cannot be nil")
	})

	t.Run("valid db returns manager", func(t *testing.T) {
		t.Parallel()

		mgr, err := uow.NewManager(newGormDB(t), func(tx *gorm.DB) repos { return repos{DB: tx} })
		require.NoError(t, err)
		assert.NotNil(t, mgr)
	})
}

func TestGormManager_Begin(t *testing.T) {
	t.Parallel()

	buildCalled := false
	mgr, err := uow.NewManager(newGormDB(t), func(tx *gorm.DB) repos {
		buildCalled = true
		return repos{DB: tx}
	})
	require.NoError(t, err)

	u, err := mgr.Begin(context.Background())
	require.NoError(t, err)
	require.NotNil(t, u)
	assert.True(t, buildCalled)

	_ = u.Rollback(context.Background())
}

func TestGormUoW_Repositories(t *testing.T) {
	t.Parallel()

	var capturedTx *gorm.DB
	mgr, err := uow.NewManager(newGormDB(t), func(tx *gorm.DB) repos {
		capturedTx = tx
		return repos{DB: tx}
	})
	require.NoError(t, err)

	u, err := mgr.Begin(context.Background())
	require.NoError(t, err)

	assert.Equal(t, capturedTx, u.Repositories().DB)

	_ = u.Rollback(context.Background())
}

func TestGormUoW_Commit(t *testing.T) {
	t.Parallel()

	gormDB := newGormDB(t)
	require.NoError(t, gormDB.Exec(`CREATE TABLE uow_commit_test (val INT)`).Error)
	t.Cleanup(func() { _ = gormDB.Exec(`DROP TABLE IF EXISTS uow_commit_test`).Error })

	mgr, err := uow.NewManager(gormDB, func(tx *gorm.DB) repos { return repos{DB: tx} })
	require.NoError(t, err)

	u, err := mgr.Begin(context.Background())
	require.NoError(t, err)

	require.NoError(t, u.Repositories().DB.Exec(`INSERT INTO uow_commit_test VALUES (42)`).Error)
	require.NoError(t, u.Commit(context.Background()))

	var count int64
	gormDB.Raw(`SELECT COUNT(*) FROM uow_commit_test WHERE val = 42`).Scan(&count)
	assert.Equal(t, int64(1), count)
}

func TestGormUoW_Rollback(t *testing.T) {
	t.Parallel()

	gormDB := newGormDB(t)
	require.NoError(t, gormDB.Exec(`CREATE TABLE uow_rollback_test (val INT)`).Error)
	t.Cleanup(func() { _ = gormDB.Exec(`DROP TABLE IF EXISTS uow_rollback_test`).Error })

	mgr, err := uow.NewManager(gormDB, func(tx *gorm.DB) repos { return repos{DB: tx} })
	require.NoError(t, err)

	u, err := mgr.Begin(context.Background())
	require.NoError(t, err)

	require.NoError(t, u.Repositories().DB.Exec(`INSERT INTO uow_rollback_test VALUES (99)`).Error)
	require.NoError(t, u.Rollback(context.Background()))

	var count int64
	gormDB.Raw(`SELECT COUNT(*) FROM uow_rollback_test WHERE val = 99`).Scan(&count)
	assert.Equal(t, int64(0), count)
}
