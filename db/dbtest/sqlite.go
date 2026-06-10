package dbtest

import (
	"context"
	"fmt"
	"io/fs"
	"sync/atomic"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/brpaz/lib-go/db"
	"github.com/brpaz/lib-go/db/migrator"
)

// sqliteSeq ensures every [NewSQLite] call gets its own isolated in-memory
// database, even when multiple tests run in parallel.
var sqliteSeq atomic.Uint64

// Options configures [NewSQLite].
type Options struct {
	MigrationsFS fs.FS
}

// Option is a functional option for [NewSQLite].
type Option func(*Options)

// WithMigrations runs all pending migrations from fsys against the database
// before [NewSQLite] returns.
func WithMigrations(fsys fs.FS) Option {
	return func(o *Options) { o.MigrationsFS = fsys }
}

// NewSQLite opens a fresh, isolated in-memory SQLite database for use in
// tests. Each call returns its own database, safe for use in parallel tests.
//
// SQLite in-memory databases only persist for as long as a connection is
// open, so the returned *gorm.DB is configured with a single connection in
// its pool to keep the schema and data visible across queries.
func NewSQLite(opts ...Option) (*gorm.DB, error) {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}

	dsn := fmt.Sprintf("file:dbtest_%d?mode=memory&cache=shared", sqliteSeq.Add(1))

	gdb, err := db.NewConnection(sqlite.Open(dsn), db.WithMaxOpenConns(1))
	if err != nil {
		return nil, fmt.Errorf("dbtest: open sqlite connection: %w", err)
	}

	if o.MigrationsFS != nil {
		if err := migrate(gdb, o.MigrationsFS); err != nil {
			return nil, err
		}
	}

	return gdb, nil
}

// migrate runs all pending migrations from fsys against gdb's underlying connection.
func migrate(gdb *gorm.DB, fsys fs.FS) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return fmt.Errorf("dbtest: retrieve sql.DB: %w", err)
	}

	m, err := migrator.NewMigrator(
		sqlDB,
		fsys,
		migrator.WithDialect(migrator.DialectSqlite),
	)
	if err != nil {
		return fmt.Errorf("dbtest: create migrator: %w", err)
	}

	if err := m.Migrate(context.Background()); err != nil {
		return fmt.Errorf("dbtest: run migrations: %w", err)
	}

	return nil
}
