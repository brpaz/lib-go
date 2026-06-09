package middleware_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/logging"
	"github.com/brpaz/lib-go/logging/logtest"
	"github.com/brpaz/lib-go/middleware"
)

func TestRequestLogger(t *testing.T) {
	t.Parallel()

	t.Run("logs completion attributes", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		handler := middleware.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))

		req := httptest.NewRequest(http.MethodPost, "/users", nil)
		req.Header.Set("User-Agent", "test-client/1.0")
		handler.ServeHTTP(httptest.NewRecorder(), req)

		entry := rec.Last()
		assert.Equal(t, "request completed", entry["message"])
		assert.Equal(t, "POST", entry["http_method"])
		assert.Equal(t, "/users", entry["http_path"])
		assert.Equal(t, "test-client/1.0", entry["http_user_agent"])
		assert.Equal(t, float64(201), entry["http_status"])
		assert.NotNil(t, entry["duration_ms"])
	})

	t.Run("defaults status to 200 when WriteHeader not called", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		handler := middleware.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		assert.Equal(t, float64(200), rec.Last()["http_status"])
	})
}

func TestRequestLogger_ContextInjection(t *testing.T) {
	t.Parallel()

	t.Run("injects logger with request attributes", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		handler := middleware.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logging.FromContext(r.Context()).Info(r.Context(), "inside handler")
		}))

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		all := rec.All()
		require.Len(t, all, 2)
		inner := all[0]
		assert.Equal(t, "inside handler", inner["message"])
		assert.Equal(t, "GET", inner["http_method"])
		assert.Equal(t, "/ping", inner["http_path"])
	})

	t.Run("preserves existing context values", func(t *testing.T) {
		t.Parallel()

		type keyType struct{}
		logger, _ := logtest.New(t)

		var gotCtx context.Context
		handler := middleware.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotCtx = r.Context()
		}))

		ctx := context.WithValue(context.Background(), keyType{}, "sentinel")
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		assert.Equal(t, "sentinel", gotCtx.Value(keyType{}))
	})
}

func TestRequestLogger_RequestBodyLogging(t *testing.T) {
	t.Parallel()

	t.Run("captures body and restores it for handler", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		var receivedBody string
		handler := middleware.RequestLogger(logger, middleware.WithRequestBodyLogging(1024))(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				receivedBody = string(data)
			}),
		)

		req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"name":"alice"}`))
		handler.ServeHTTP(httptest.NewRecorder(), req)

		entry := rec.Last()
		assert.Equal(t, `{"name":"alice"}`, entry["http_request_body"])
		assert.Nil(t, entry["http_request_body_truncated"])
		assert.Equal(t, `{"name":"alice"}`, receivedBody)
	})

	t.Run("truncates body exceeding limit", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		handler := middleware.RequestLogger(logger, middleware.WithRequestBodyLogging(5))(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.ReadAll(r.Body) //nolint:errcheck
			}),
		)

		req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("hello world"))
		handler.ServeHTTP(httptest.NewRecorder(), req)

		entry := rec.Last()
		assert.Equal(t, "hello", entry["http_request_body"])
		assert.Equal(t, true, entry["http_request_body_truncated"])
	})
}

func TestRequestLogger_ResponseBodyLogging(t *testing.T) {
	t.Parallel()

	t.Run("captures response body", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		handler := middleware.RequestLogger(logger, middleware.WithResponseBodyLogging(1024))(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"id":"u-1"}`)) //nolint:errcheck
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		entry := rec.Last()
		assert.Equal(t, `{"id":"u-1"}`, entry["http_response_body"])
		assert.Nil(t, entry["http_response_body_truncated"])
	})

	t.Run("truncates body exceeding limit", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		handler := middleware.RequestLogger(logger, middleware.WithResponseBodyLogging(5))(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello world")) //nolint:errcheck
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/big", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		entry := rec.Last()
		assert.Equal(t, "hello", entry["http_response_body"])
		assert.Equal(t, true, entry["http_response_body_truncated"])
	})

	t.Run("captures both request and response body", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		handler := middleware.RequestLogger(logger,
			middleware.WithRequestBodyLogging(1024),
			middleware.WithResponseBodyLogging(1024),
		)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.ReadAll(r.Body)     //nolint:errcheck
			w.Write([]byte("ok")) //nolint:errcheck
		}))

		req := httptest.NewRequest(http.MethodPost, "/process", strings.NewReader("input"))
		handler.ServeHTTP(httptest.NewRecorder(), req)

		entry := rec.Last()
		assert.Equal(t, "input", entry["http_request_body"])
		assert.Equal(t, "ok", entry["http_response_body"])
	})
}
