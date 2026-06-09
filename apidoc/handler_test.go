package apidoc_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/apidoc"
)

const testSpec = "openapi: 3.0.0\ninfo:\n  title: Test\n"

func newHandler(t *testing.T, opts ...apidoc.Option) *apidoc.Handler {
	t.Helper()
	h, err := apidoc.New(append([]apidoc.Option{apidoc.WithSpec([]byte(testSpec))}, opts...)...)
	require.NoError(t, err)
	return h
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestNew_MissingSpec(t *testing.T) {
	t.Parallel()
	_, err := apidoc.New()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "spec is required")
}

func TestServeHTTP_SpecRoute(t *testing.T) {
	t.Parallel()

	w := get(newHandler(t), "/openapi.yml")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/yaml", w.Header().Get("Content-Type"))
	assert.Equal(t, testSpec, w.Body.String())
}

func TestServeHTTP_UIRoute_Disabled(t *testing.T) {
	t.Parallel()

	w := get(newHandler(t), "/")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServeHTTP_UIRoute_Enabled(t *testing.T) {
	t.Parallel()

	w := get(newHandler(t, apidoc.WithScalarUI()), "/")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `data-url="./openapi.yml"`)
	assert.Contains(t, w.Body.String(), "cdn.jsdelivr.net")
}

func TestServeHTTP_UnknownPath(t *testing.T) {
	t.Parallel()

	w := get(newHandler(t), "/unknown")
	assert.Equal(t, http.StatusNotFound, w.Code)
}
