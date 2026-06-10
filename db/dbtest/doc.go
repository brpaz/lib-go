// Package dbtest provides lightweight, driver-agnostic test helpers for code
// built on [github.com/brpaz/lib-go/db].
//
// # SQLite
//
// [NewSQLite] opens an isolated in-memory SQLite database, ready to use as a
// *gorm.DB. No external services or containers required:
//
//	func TestSomething(t *testing.T) {
//	    gdb, err := dbtest.NewSQLite()
//	    require.NoError(t, err)
//	}
//
// Pass [WithMigrations] with an [io/fs.FS] containing goose SQL migration
// files to apply them before [NewSQLite] returns:
//
//	//go:embed migrations/*.sql
//	var migrationsFS embed.FS
//
//	gdb, err := dbtest.NewSQLite(dbtest.WithMigrations(migrationsFS))
//
// # Transactions
//
// [WithTx] opens a transaction on an existing *gorm.DB and rolls it back when
// the test ends, giving each test a clean slate:
//
//	func TestCreateUser(t *testing.T) {
//	    tx := dbtest.WithTx(t, gormDB)
//	    svc := user.NewService(tx)
//	    // any writes are rolled back when t ends
//	}
//
// # PostgreSQL
//
// For tests that need a real PostgreSQL instance, see the separate module
// [github.com/brpaz/lib-go/db/dbtest/postgres], which spins up ephemeral
// containers via Testcontainers. It is a separate module so that consumers of
// this package don't pull in Testcontainers and the Postgres driver.
package dbtest
