package checks_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/health"
	"github.com/brpaz/lib-go/health/checks"
)

func TestNewURLCheck(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)

	check, err := checks.NewURLCheck("test", checks.WithURLCheckURL(*u))

	assert.NoError(t, err)
	assert.IsType(t, &checks.URLCheck{}, check)
}

func TestURLCheck_Check(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)

		check, _ := checks.NewURLCheck("test", checks.WithURLCheckURL(*u))

		result := check.Check(context.Background())

		assert.Equal(t, health.StatusPass, result.Status)
		assert.NoError(t, result.Error)
	})

	t.Run("Failure_OnUnexpectedStatusCode", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)

		check, _ := checks.NewURLCheck("test", checks.WithURLCheckURL(*u))

		result := check.Check(context.Background())

		assert.Equal(t, health.StatusFail, result.Status)
		assert.Error(t, result.Error)
		assert.Equal(t, "unexpected response. status: 500 body: Internal Server Error", result.Message)
	})

	t.Run("Failure_OnRequestTimeout", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
		}))
		defer ts.Close()

		u, _ := url.Parse(ts.URL)

		check, _ := checks.NewURLCheck("test", checks.WithURLCheckURL(*u), checks.WithURLCheckTimeout(1))

		result := check.Check(context.Background())

		assert.Equal(t, health.StatusFail, result.Status)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "context deadline exceeded")
	})
}
