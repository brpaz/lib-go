package db_test

import (
	"context"
	"os"
	"testing"

	dbtest "github.com/brpaz/lib-go/storage/db/testutil"
)

var dbInstance *dbtest.TestPgContainer

func setupTestDb(ctx context.Context) (*dbtest.TestPgContainer, error) {
	return dbtest.InitPgTestContainer(ctx)
}

// TestMain is the entry point for running tests in this package.
// It can be used to initialize global resources that are needed across tests.
func TestMain(m *testing.M) {
	ctx := context.Background()
	db, err := setupTestDb(ctx)
	if err != nil {
		panic(err)
	}

	dbInstance = db

	// Execute the tests
	exitCode := m.Run()

	_ = dbInstance.Stop(ctx)

	// Exit with the test result
	os.Exit(exitCode)
}
