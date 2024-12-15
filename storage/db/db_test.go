package db_test

import (
	"context"
	"os"
	"testing"

	dbtestutil "github.com/brpaz/lib-go/storage/db/testutil"
)

var dbInstance *dbtestutil.TestPgContainer

func setupTestDb(ctx context.Context) (*dbtestutil.TestPgContainer, error) {
	return dbtestutil.InitPgTestContainer(ctx)
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
