// Package postgres provides test helpers for spinning up ephemeral PostgreSQL
// instances using Testcontainers.
//
// This is a separate module from [github.com/brpaz/lib-go/db/dbtest] so that
// projects only needing the lightweight [github.com/brpaz/lib-go/db/dbtest.NewSQLite]
// helper don't pull in Testcontainers and the Postgres driver.
//
// # Usage
//
// For a per-test container:
//
//	func TestSomething(t *testing.T) {
//	    pg, err := postgres.NewTestContainer(context.Background())
//	    require.NoError(t, err)
//	    pg.Cleanup(t) // terminates the container when t ends
//	}
//
// For a package-scoped container shared across all tests, set a package-level
// variable from TestMain and use [Container.NewIsolatedDB] per test:
//
//	var testPG *postgres.Container
//
//	func TestMain(m *testing.M) {
//	    var err error
//	    testPG, err = postgres.NewTestContainer(context.Background())
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
// # Migrations
//
// Pass [WithAutoMigrate] together with [WithMigrationsFS] (an [io/fs.FS]
// containing goose SQL migration files) to apply them to the container's
// default database before [NewTestContainer] returns. Use [WithMigrationsDir]
// if the SQL files live in a sub-directory of the FS:
//
//	//go:embed migrations/*.sql
//	var migrationsFS embed.FS
//
//	pg, err := postgres.NewTestContainer(context.Background(),
//	    postgres.WithAutoMigrate(),
//	    postgres.WithMigrationsFS(migrationsFS),
//	    postgres.WithMigrationsDir("migrations"),
//	)
package postgres
