package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	libhttp "github.com/brpaz/lib-go/http"
)

func TestOk(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rr := httptest.NewRecorder()

		body := map[string]string{"message": "success"}

		libhttp.Ok(rr, body)

		assert.Equal(t, http.StatusOK, rr.Code)

		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		// Check the response body
		expectedBody, _ := json.Marshal(body)
		assert.JSONEq(t, string(expectedBody), rr.Body.String())
	})

	t.Run("error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		body := make(chan int)

		libhttp.Ok(rr, body)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "json: unsupported type")
	})
}

func TestNoContent(t *testing.T) {
	rr := httptest.NewRecorder()

	libhttp.NoContent(rr)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, 0, rr.Body.Len())
}
