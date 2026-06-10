package dbtest

import (
	"fmt"
	"sync/atomic"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// sqliteSeq ensures every [NewSQLite] call gets its own isolated in-memory
// database, even when multiple tests run in parallel.
var sqliteSeq atomic.Uint64

// SQLiteOptions configures [NewSQLite].
type SQLiteOptions struct {
}

// SQLiteOption is a functional option for [NewSQLite].
type SQLiteOption func(*SQLiteOptions)

// NewSQLite opens a fresh, isolated in-memory SQLite database for use in
// tests. Each call returns its own database, safe for use in parallel tests.
//
// SQLite in-memory databases only persist for as long as a connection is
// open, so the returned *gorm.DB is configured with a single connection in
// its pool to keep the schema and data visible across queries.
func NewSQLite(opts ...SQLiteOption) (*gorm.DB, error) {
	o := &SQLiteOptions{}
	for _, opt := range opts {
		opt(o)
	}

	dsn := fmt.Sprintf("file:dbtest_%d?mode=memory&cache=shared", sqliteSeq.Add(1))

	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("dbtest: open sqlite connection: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("dbtest: retrieve sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	return gdb, nil
}
