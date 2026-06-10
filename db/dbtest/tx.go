package dbtest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// WithTx opens a database transaction on db and returns it as a *gorm.DB.
// The transaction is automatically rolled back when t ends, so each test
// starts with a clean slate without needing to reset or drop tables.
//
// The returned *gorm.DB is a drop-in replacement for a regular *gorm.DB and
// can be passed directly to application services and repositories under test.
//
//	func TestCreateUser(t *testing.T) {
//	    tx := dbtest.WithTx(t, gormDB)
//	    svc := user.NewService(tx)
//	    // any writes are rolled back when t ends
//	}
func WithTx(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()

	tx := db.Begin()
	require.NoError(t, tx.Error, "dbtest: failed to begin transaction")

	t.Cleanup(func() {
		_ = tx.Rollback()
	})

	return tx
}
