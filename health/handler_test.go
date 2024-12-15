package health_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/health"
	"github.com/brpaz/lib-go/health/checks"
)

func loadFileFromTestdata(t *testing.T, filePath string) string {
	t.Helper()

	data, err := os.ReadFile(filePath)
	require.NoError(t, err, "Failed to read JSON file")

	return string(data)
}

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	t.Run("HealthyResponse", func(t *testing.T) {
		t.Parallel()
		expectedResponse := loadFileFromTestdata(t, "testdata/health_pass_response.json")

		check1 := checks.NewStubCheck("check:1", true)
		check2 := checks.NewStubCheck("check:2", true)

		service := setupTestService(t, check1, check2)

		handler := health.Handler(service)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusOK, w.Code)

		assert.JSONEq(t, expectedResponse, w.Body.String())
	})

	t.Run("UnhealthyResponse", func(t *testing.T) {
		t.Parallel()
		expectedResponse := loadFileFromTestdata(t, "testdata/health_fail_response.json")

		check1 := checks.NewStubCheck("check:1", false)

		service := setupTestService(t, check1)

		handler := health.Handler(service)

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.JSONEq(t, expectedResponse, w.Body.String())
	})
}
