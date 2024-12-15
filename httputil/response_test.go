package httputil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/httputil"
)

func TestJSON(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()

		body := map[string]string{"message": "success"}

		httputil.JSON(rr, http.StatusOK, body)

		assert.Equal(t, http.StatusOK, rr.Code)

		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		// Check the response body
		expectedBody, _ := json.Marshal(body)
		assert.JSONEq(t, string(expectedBody), rr.Body.String())
	})

	t.Run("error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		body := make(chan int)

		httputil.JSON(rr, http.StatusAccepted, body)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "json: unsupported type")
	})
}
