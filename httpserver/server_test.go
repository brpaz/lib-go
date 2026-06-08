package httpserver_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/httpserver"
)

// freePort asks the OS for an available TCP port on localhost.
func freePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	return l.Addr().(*net.TCPAddr).Port
}

// waitForListening blocks until something accepts TCP connections on port.
func waitForListening(t *testing.T, port int) {
	t.Helper()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()

		return true
	}, 2*time.Second, 10*time.Millisecond, "server did not start listening on %s", addr)
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("succeeds with valid handler and port", func(t *testing.T) {
		t.Parallel()

		srv, err := httpserver.New(
			httpserver.WithHandler(http.NotFoundHandler()),
			httpserver.WithPort(8081),
		)
		require.NoError(t, err)
		require.NotNil(t, srv)
	})

	t.Run("fails when handler is nil", func(t *testing.T) {
		t.Parallel()

		srv, err := httpserver.New(httpserver.WithPort(8081))
		require.Error(t, err)
		assert.Nil(t, srv)
		assert.Contains(t, err.Error(), "handler is required")
	})

	invalidPorts := []struct {
		name string
		port int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too high", 65536},
	}
	for _, tc := range invalidPorts {
		t.Run("fails with invalid port "+tc.name, func(t *testing.T) {
			t.Parallel()

			srv, err := httpserver.New(
				httpserver.WithHandler(http.NotFoundHandler()),
				httpserver.WithPort(tc.port),
			)
			require.Error(t, err)
			assert.Nil(t, srv)
			assert.Contains(t, err.Error(), "invalid port")
		})
	}

	t.Run("applies default port 8080 when WithPort is not called", func(t *testing.T) {
		t.Parallel()

		srv, err := httpserver.New(
			httpserver.WithHandler(http.NotFoundHandler()),
		)
		require.NoError(t, err)
		require.NotNil(t, srv)
		assert.Equal(t, ":8080", srv.Addr())
	})

	t.Run("applies custom timeouts", func(t *testing.T) {
		t.Parallel()

		srv, err := httpserver.New(
			httpserver.WithHandler(http.NotFoundHandler()),
			httpserver.WithPort(9090),
			httpserver.WithReadTimeout(1*time.Second),
			httpserver.WithWriteTimeout(2*time.Second),
			httpserver.WithIdleTimeout(3*time.Second),
		)
		require.NoError(t, err)
		require.NotNil(t, srv)
		assert.Equal(t, ":9090", srv.Addr())
	})
}

func TestServer_Addr(t *testing.T) {
	t.Parallel()

	srv, err := httpserver.New(
		httpserver.WithHandler(http.NotFoundHandler()),
		httpserver.WithPort(7070),
	)
	require.NoError(t, err)
	assert.Equal(t, ":7070", srv.Addr())
}

func TestServer_StartAndShutdown(t *testing.T) {
	t.Parallel()

	// Find a free port.
	port := freePort(t)

	srv, err := httpserver.New(
		httpserver.WithHandler(http.NotFoundHandler()),
		httpserver.WithPort(port),
	)
	require.NoError(t, err)

	startErr := make(chan error, 1)
	go func() {
		startErr <- srv.Start()
	}()

	// Give the server time to bind.
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, srv.Shutdown(ctx))

	err = <-startErr
	assert.True(t, err == nil || errors.Is(err, http.ErrServerClosed),
		"expected nil or http.ErrServerClosed, got: %v", err)
}

func TestServer_Run(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when context is cancelled before the server binds", func(t *testing.T) {
		t.Parallel()

		srv, err := httpserver.New(
			httpserver.WithHandler(http.NotFoundHandler()),
			httpserver.WithPort(freePort(t)),
		)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		assert.NoError(t, srv.Run(ctx))
	})

	t.Run("returns nil on graceful shutdown when an active connection drains in time", func(t *testing.T) {
		t.Parallel()

		started := make(chan struct{})

		// Sleeps briefly so it's still in flight when shutdown starts, but
		// finishes well within the shutdown timeout — letting it drain.
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			time.Sleep(150 * time.Millisecond)
		})

		port := freePort(t)
		srv, err := httpserver.New(
			httpserver.WithHandler(handler),
			httpserver.WithPort(port),
			httpserver.WithShutdownTimeout(5*time.Second),
		)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		runErr := make(chan error, 1)
		go func() { runErr <- srv.Run(ctx) }()

		waitForListening(t, port)

		getDone := make(chan struct{})
		go func() {
			defer close(getDone)

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port)) //nolint:noctx,gosec // test client; URL is constructed from a local port.
			if err == nil {
				_ = resp.Body.Close()
			}
		}()

		<-started
		cancel()

		select {
		case err := <-runErr:
			assert.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("Run did not return after context cancellation")
		}

		<-getDone
	})

	t.Run("returns the start error when the port is already in use", func(t *testing.T) {
		t.Parallel()

		port := freePort(t)

		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		require.NoError(t, err)
		defer func() { _ = l.Close() }()

		srv, err := httpserver.New(
			httpserver.WithHandler(http.NotFoundHandler()),
			httpserver.WithPort(port),
		)
		require.NoError(t, err)

		err = srv.Run(context.Background())
		require.Error(t, err)
		assert.NotErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("returns a wrapped error when shutdown exceeds the configured timeout", func(t *testing.T) {
		t.Parallel()

		release := make(chan struct{})

		blocked := make(chan struct{})
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(blocked)
			<-release
		})

		port := freePort(t)
		srv, err := httpserver.New(
			httpserver.WithHandler(handler),
			httpserver.WithPort(port),
			httpserver.WithShutdownTimeout(50*time.Millisecond),
		)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		runErr := make(chan error, 1)
		go func() { runErr <- srv.Run(ctx) }()

		waitForListening(t, port)

		getDone := make(chan struct{})
		go func() {
			defer close(getDone)

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port)) //nolint:noctx,gosec // test client; URL is constructed from a local port.
			if err == nil {
				_ = resp.Body.Close()
			}
		}()

		<-blocked
		cancel()

		select {
		case err := <-runErr:
			require.Error(t, err)
			assert.ErrorIs(t, err, context.DeadlineExceeded)
		case <-time.After(5 * time.Second):
			t.Fatal("Run did not return after shutdown timeout")
		}

		close(release)
		<-getDone
	})
}
