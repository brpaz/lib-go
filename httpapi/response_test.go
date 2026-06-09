package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/httpapi"
)

type testPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestOk(t *testing.T) {
	t.Parallel()

	t.Run("sets status 200", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Ok(w, testPayload{Name: "alice", Age: 30})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("sets Content-Type application/json", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Ok(w, testPayload{})
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("encodes body as JSON", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Ok(w, testPayload{Name: "alice", Age: 30})
		var got testPayload
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, testPayload{Name: "alice", Age: 30}, got)
	})
}

func TestCreated(t *testing.T) {
	t.Parallel()

	t.Run("sets status 201", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Created(w, testPayload{Name: "bob"})
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("sets Content-Type application/json", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Created(w, testPayload{})
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("encodes body as JSON", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Created(w, testPayload{Name: "bob", Age: 25})
		var got testPayload
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, testPayload{Name: "bob", Age: 25}, got)
	})
}

func TestAccepted(t *testing.T) {
	t.Parallel()

	t.Run("sets status 202", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Accepted(w, testPayload{Name: "carol"})
		assert.Equal(t, http.StatusAccepted, w.Code)
	})

	t.Run("sets Content-Type application/json", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Accepted(w, testPayload{})
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("encodes body as JSON", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Accepted(w, testPayload{Name: "carol", Age: 40})
		var got testPayload
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, testPayload{Name: "carol", Age: 40}, got)
	})
}

func TestError(t *testing.T) {
	t.Parallel()

	t.Run("sets given status code", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Error(w, http.StatusNotFound, "not_found", "resource not found")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("sets Content-Type application/json", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Error(w, http.StatusNotFound, "not_found", "resource not found")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})

	t.Run("encodes code and message", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.Error(w, http.StatusNotFound, "not_found", "resource not found")
		var got httpapi.ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, "not_found", got.Code)
		assert.Equal(t, "resource not found", got.Message)
	})
}

func TestBadRequest(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httpapi.BadRequest(w, testPayload{Name: "err"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	var got testPayload
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Equal(t, testPayload{Name: "err"}, got)
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httpapi.Unauthorized(w, testPayload{Name: "err"})
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestForbidden(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httpapi.Forbidden(w, testPayload{Name: "err"})
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestNotFound(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httpapi.NotFound(w, testPayload{Name: "err"})
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestConflict(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httpapi.Conflict(w, testPayload{Name: "err"})
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnprocessableEntity(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httpapi.UnprocessableEntity(w, testPayload{Name: "err"})
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestTooManyRequests(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httpapi.TooManyRequests(w, testPayload{Name: "err"})
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestInternalServerError(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	httpapi.InternalServerError(w, testPayload{Name: "err"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestNoContent(t *testing.T) {
	t.Parallel()

	t.Run("sets status 204", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.NoContent(w)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("writes no body", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		httpapi.NoContent(w)
		assert.Empty(t, w.Body.String())
	})
}
