package spa_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/brpaz/lib-go/spa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFS() fs.FS {
	return fstest.MapFS{
		"index.html":        {Data: []byte(`<html>__ENV__</html>`)},
		"assets/app.js":     {Data: []byte(`console.log("app")`)},
		"assets/style.css":  {Data: []byte(`body{}`)},
		"custom/shell.html": {Data: []byte(`<html>shell</html>`)},
	}
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandler_StaticAsset(t *testing.T) {
	h := spa.Handler(testFS())
	rec := get(h, "/assets/app.js")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "console.log")
}

func TestHandler_UnknownPathFallsBackToIndex(t *testing.T) {
	h := spa.Handler(testFS())
	for _, path := range []string{"/", "/dashboard", "/settings/profile"} {
		rec := get(h, path)
		require.Equal(t, http.StatusOK, rec.Code, "path: %s", path)
		assert.Contains(t, rec.Body.String(), "<html>", "path: %s", path)
	}
}

func TestHandler_IndexCacheControl(t *testing.T) {
	h := spa.Handler(testFS())
	rec := get(h, "/")
	assert.Equal(t, "no-cache, no-store, must-revalidate", rec.Header().Get("Cache-Control"))
}

func TestHandler_WithIndexCacheControl(t *testing.T) {
	h := spa.Handler(testFS(), spa.WithIndexCacheControl("no-cache"))
	rec := get(h, "/unknown")
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
}

func TestHandler_WithEnvVars(t *testing.T) {
	vars := map[string]any{"API_URL": "https://api.example.com"}
	h := spa.Handler(testFS(), spa.WithEnvVars("__ENV__", vars))
	rec := get(h, "/")
	body := rec.Body.String()
	assert.Contains(t, body, `"API_URL"`)
	assert.NotContains(t, body, "__ENV__")
}

func TestHandler_WithIndexFile(t *testing.T) {
	h := spa.Handler(testFS(), spa.WithIndexFile("custom/shell.html"))
	rec := get(h, "/unknown-path")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "shell")
}

func TestHandler_WithNotFoundHandler(t *testing.T) {
	h := spa.Handler(testFS(), spa.WithNotFoundHandler(http.NotFoundHandler()))
	rec := get(h, "/does-not-exist")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_WithStaticCacheControl(t *testing.T) {
	const cc = "public, max-age=31536000, immutable"
	h := spa.Handler(testFS(), spa.WithStaticCacheControl(cc))
	rec := get(h, "/assets/app.js")
	assert.Equal(t, cc, rec.Header().Get("Cache-Control"))
}

func TestHandler_StaticNoCacheControlByDefault(t *testing.T) {
	h := spa.Handler(testFS())
	rec := get(h, "/assets/app.js")
	assert.Empty(t, rec.Header().Get("Cache-Control"))
}

func TestHandler_ContentType(t *testing.T) {
	h := spa.Handler(testFS())
	rec := get(h, "/")
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
}
