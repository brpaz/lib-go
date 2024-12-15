package checks_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/health"
	"github.com/brpaz/lib-go/health/checks"
)

func setupTestStubCheck(t *testing.T, result bool) *checks.StubCheck {
	t.Helper()
	return checks.NewStubCheck("stub", result)
}

func TestStubCheck_GetName(t *testing.T) {
	t.Parallel()
	check := setupTestStubCheck(t, true)
	assert.Equal(t, "stub", check.GetName())
}

func TestStubCheck_Check(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		check := setupTestStubCheck(t, true)
		result := check.Check(context.Background())
		assert.Equal(t, health.StatusPass, result.Status)
		assert.NoError(t, result.Error)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Parallel()
		check := setupTestStubCheck(t, false)
		result := check.Check(context.Background())

		assert.Equal(t, health.StatusFail, result.Status)
		assert.Error(t, result.Error)
		assert.Equal(t, "Stub check failed", result.Message)
	})
}
