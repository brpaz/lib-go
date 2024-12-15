package checks_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/health"
	"github.com/brpaz/lib-go/health/checks"
	dbtestutil "github.com/brpaz/lib-go/storage/db/testutil"
)

func TestDBCheck(t *testing.T) {
	ctx := context.Background()
	dbInstance, err := dbtestutil.InitPgTestContainer(ctx)
	require.NoError(t, err)

	dbConn, err := dbInstance.GetConnection(ctx)
	require.NoError(t, err)

	check := checks.NewDBCheck("test", dbConn)

	result := check.Check(ctx)

	assert.Equal(t, health.StatusPass, result.Status)
}
