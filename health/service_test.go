package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/health"
	"github.com/brpaz/lib-go/health/checks"
	"github.com/brpaz/lib-go/timeutil"
)

// Helper function to setup a test service with the provided checks.
func setupTestService(t *testing.T, checks ...health.Checker) *health.Service {
	t.Helper()
	return health.New(
		health.WithName("test-service"),
		health.WithDescription("Test Service"),
		health.WithVersion("1.0.0"),
		health.WithRevision("abc123"),
		health.WithChecks(checks...),
		health.WithClock(timeutil.NewMockClock(time.Date(2024, time.June, 6, 13, 5, 10, 0, time.UTC))),
	)
}

func TestNew(t *testing.T) {
	t.Parallel()

	service := setupTestService(t, checks.NewStubCheck("stub", true))

	assert.NotNil(t, service)
	assert.Equal(t, "test-service", service.Name)
	assert.Equal(t, "1.0.0", service.Version)
	assert.Equal(t, "abc123", service.Revision)
	assert.Len(t, service.Checks, 1)
}

func TestService_AddCheck(t *testing.T) {
	t.Parallel()

	service := setupTestService(t)
	service.AddCheck(checks.NewStubCheck("stub", true))

	assert.Len(t, service.Checks, 1)
	assert.IsType(t, &checks.StubCheck{}, service.Checks[0])
}

func TestService_Execute(t *testing.T) {
	t.Parallel()

	t.Run("WithPassHealthStatus", func(t *testing.T) {
		t.Parallel()
		service := setupTestService(t, checks.NewStubCheck("stub", true))

		result := service.Execute(context.Background())

		assert.Equal(t, health.StatusPass, result.Status)
		assert.Len(t, result.Checks, 1)

		stubCheckResult := result.Checks["stub"]

		require.NotNil(t, stubCheckResult)
		assert.Equal(t, health.StatusPass, stubCheckResult.Status)
		assert.NoError(t, stubCheckResult.Error)
	})

	t.Run("WithFailHealthStatus", func(t *testing.T) {
		t.Parallel()
		service := setupTestService(t, checks.NewStubCheck("stub", false))

		result := service.Execute(context.Background())

		assert.Equal(t, health.StatusFail, result.Status)
		assert.Len(t, result.Checks, 1)

		stubCheckResult := result.Checks["stub"]

		require.NotNil(t, stubCheckResult)
		assert.Equal(t, health.StatusFail, stubCheckResult.Status)
		assert.Error(t, stubCheckResult.Error)
		assert.Equal(t, "stub check failed", stubCheckResult.Error.Error())
	})
}
