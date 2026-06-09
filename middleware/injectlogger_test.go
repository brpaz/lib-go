package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/logging"
	"github.com/brpaz/lib-go/logging/logtest"
	"github.com/brpaz/lib-go/middleware"
)

func TestInjectLogger(t *testing.T) {
	t.Parallel()

	t.Run("logger is retrievable from context", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)
		var ctxLogger *logging.Logger

		handler := middleware.InjectLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxLogger = logging.FromContext(r.Context())
			ctxLogger.Info(r.Context(), "handler called")
		}))

		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		require.NotNil(t, ctxLogger)
		require.Equal(t, 1, rec.Len())
		assert.Equal(t, "handler called", rec.Last()["message"])
	})

	t.Run("does not log a completion entry", func(t *testing.T) {
		t.Parallel()

		logger, rec := logtest.New(t)

		handler := middleware.InjectLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, 0, rec.Len())
	})
}
