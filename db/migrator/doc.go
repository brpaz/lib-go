// Package migrator runs database schema migrations using goose.
//
// # Basic usage
//
// Pass an *sql.DB and an fs.FS containing SQL files under a "migrations/"
// sub-path. Callers typically embed the files and pass the resulting embed.FS:
//
//	//go:embed migrations/*.sql
//	var migrationsFS embed.FS
//
//	m, err := migrator.NewMigrator(db, migrationsFS)
//
// [Migrator.Migrate] applies all pending UP migrations.
// [Migrator.Rollback] reverts the last applied migration.
// [Migrator.Status] logs the current applied/pending state.
//
// # Logging
//
// By default the migrator discards all output. Supply a [Logger] via
// [WithLogger] to route migration output to your preferred sink:
//
//	m, err := migrator.NewMigrator(db, fsys,
//	    migrator.WithLogger(migrator.NewLibLogger(appLogger)),
//	)
//
// [NewLibLogger] wraps a [github.com/brpaz/lib-go/logging.Logger]. [NewNopLogger]
// discards output. Any value that implements [Logger] (Printf + Fatalf) is
// accepted, so adapters for zap, logrus, or other libraries can be plugged in
// directly.
//
// # Concurrency
//
// goose uses package-level global state (base FS, dialect, table name).
// Avoid running multiple Migrator instances concurrently with different options.
package migrator
