// Package dbtest provides lightweight test helpers for code built on
// [github.com/brpaz/lib-go/db].
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
// [NewPostgresContainer] starts an ephemeral PostgreSQL instance via
// Testcontainers:
//
//	func TestSomething(t *testing.T) {
//	    pg, err := dbtest.NewPostgresContainer(context.Background())
//	    require.NoError(t, err)
//	    pg.Cleanup(t) // terminates the container when t ends
//	}
//
// For a package-scoped container shared across all tests, set a package-level
// variable from TestMain and use [PostgresContainer.NewIsolatedDB] per test:
//
//	var testPG *dbtest.PostgresContainer
//
//	func TestMain(m *testing.M) {
//	    var err error
//	    testPG, err = dbtest.NewPostgresContainer(context.Background())
//	    if err != nil { log.Fatal(err) }
//	    code := m.Run()
//	    testPG.Terminate(context.Background())
//	    os.Exit(code)
//	}
//
//	func TestSomething(t *testing.T) {
//	    db := testPG.NewIsolatedDB(t) // fresh database, dropped on test end
//	}
//
// Use [MigratePostgres] to apply goose SQL migrations to a database before
// running tests against it (the container's default database, or one
// returned by [PostgresContainer.NewIsolatedDB]):
//
//	//go:embed migrations/*.sql
//	var migrationsFS embed.FS
//
//	pg, err := dbtest.NewPostgresContainer(context.Background())
//	require.NoError(t, err)
//	require.NoError(t, dbtest.MigratePostgres(context.Background(), pg.DSN, migrationsFS))
//
// # SQLite
//
// [NewSQLite] opens a fresh, isolated in-memory SQLite database — no Docker
// required. Each call gets its own database, safe for parallel tests:
//
//	func TestSomething(t *testing.T) {
//	    gdb, err := dbtest.NewSQLite()
//	    require.NoError(t, err)
//	}
//
// Pass [WithSQLiteMigrations] to apply goose SQL migrations before
// [NewSQLite] returns:
//
//	//go:embed migrations/*.sql
//	var migrationsFS embed.FS
//
//	gdb, err := dbtest.NewSQLite(dbtest.WithSQLiteMigrations(migrationsFS))
package dbtest
