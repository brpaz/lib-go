package migrator_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/storage/db/migrator"
	dbtestutil "github.com/brpaz/lib-go/storage/db/testutil"
)

var dbInstance *dbtestutil.TestPgContainer

func setupTestDb(ctx context.Context) (*dbtestutil.TestPgContainer, error) {
	return dbtestutil.InitPgTestContainer(ctx)
}

func setupMigrator(t *testing.T, dbConn *sql.DB) (*migrator.GooseMigrator, error) {
	t.Helper()

	return migrator.NewGooseMigrator(
		migrator.WithGooseDB(dbConn),
		migrator.WithGooseAllowOutOfOrder(true),
		migrator.WithGooseDialect("postgres"),
		migrator.WithGooseMigrationsType("sql"),
		migrator.WithGooseSequencial(true),
		migrator.WithGooseMigrationsDir("testdata/migrations/sql"),
	)
}

// tableExists checks if a table exists in the database
func tableExists(dbConn *sql.DB, tableName string) bool {
	var exists bool

	// nolint: gosec
	query := fmt.Sprintf("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = '%s')", tableName)
	err := dbConn.QueryRow(query).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func TestNewGooseMigrator(t *testing.T) {
	dbConn, err := dbInstance.GetConnection(context.Background())
	require.NoError(t, err)

	t.Run("WithMissingDB_ReturnsError", func(t *testing.T) {
		_, err := migrator.NewGooseMigrator()
		require.Error(t, err)
		assert.IsType(t, err, migrator.ErrMissingSqlDB)
	})

	t.Run("WithValidOptions_ReturnsMigrator", func(t *testing.T) {
		m, err := setupMigrator(t, dbConn)
		require.NoError(t, err)
		assert.IsType(t, m, &migrator.GooseMigrator{})
	})

	t.Run("WithInvalidMigrationType_ReturnsError", func(t *testing.T) {
		_, err := migrator.NewGooseMigrator(
			migrator.WithGooseDB(dbConn),
			migrator.WithGooseMigrationsType("invalid"),
		)
		require.Error(t, err)
		assert.IsType(t, err, migrator.ErrInvalidMigrationsType)
	})
}

func TestGooseMigrator_Up(t *testing.T) {
	ctx := context.Background()
	dbConn, err := dbInstance.GetConnection(ctx)
	require.NoError(t, err)

	m, err := setupMigrator(t, dbConn)
	require.NoError(t, err)

	err = m.Up(ctx)
	require.NoError(t, err)

	// Assert that the migrations were applied
	assert.True(t, tableExists(dbConn, "goose_db_version"))
	assert.True(t, tableExists(dbConn, "users"))
}

func TestGooseMigrator_Down(t *testing.T) {
	ctx := context.Background()
	dbConn, err := dbInstance.GetConnection(ctx)
	require.NoError(t, err)

	m, err := setupMigrator(t, dbConn)
	require.NoError(t, err)

	err = m.Up(ctx)
	require.NoError(t, err)

	err = m.Down(ctx)
	require.NoError(t, err)

	assert.False(t, tableExists(dbConn, "users"))
}

func TestGooseMigrator_Reset(t *testing.T) {
	ctx := context.Background()
	dbConn, err := dbInstance.GetConnection(ctx)
	require.NoError(t, err)

	m, err := setupMigrator(t, dbConn)
	require.NoError(t, err)

	err = m.Up(ctx)
	require.NoError(t, err)

	err = m.Reset(ctx)
	require.NoError(t, err)

	assert.False(t, tableExists(dbConn, "users"))
}

func TestGooseMigrator_Create(t *testing.T) {
	ctx := context.Background()
	dbConn, err := dbInstance.GetConnection(ctx)
	require.NoError(t, err)

	migrationsDir, err := os.MkdirTemp("", "migrations")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(migrationsDir)
	}()

	m, err := migrator.NewGooseMigrator(
		migrator.WithGooseDB(dbConn),
		migrator.WithGooseDialect("postgres"),
		migrator.WithGooseMigrationsType("sql"),
		migrator.WithGooseSequencial(true),
		migrator.WithGooseMigrationsDir(migrationsDir),
	)
	require.NoError(t, err)

	err = m.Create(ctx, "test_migration")
	require.NoError(t, err)

	// read the migration file
	_, err = os.Stat(fmt.Sprintf("%s/%s.sql", migrationsDir, "00001_test_migration"))
	require.NoError(t, err)
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
