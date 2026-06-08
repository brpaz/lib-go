// Package httpserver provides a [Server] that wraps [net/http.Server] with
// sane defaults and graceful shutdown support.
//
// # Creating a server
//
// Build a [Server] with [New], passing a fully configured handler (e.g. a
// router). The server owns no routing logic of its own:
//
//	srv, err := httpserver.New(
//	    httpserver.WithHandler(router),
//	    httpserver.WithPort(8080),
//	)
//	if err != nil {
//	    return err
//	}
//
// [WithHandler] is the only required option. [WithPort], [WithReadTimeout],
// [WithWriteTimeout] and [WithIdleTimeout] override the package defaults.
//
// # Running and stopping
//
// Start blocks until the server stops, so run it in a goroutine and wait for
// a shutdown signal:
//
//	go func() {
//	    if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
//	        log.Fatal(err)
//	    }
//	}()
//
//	<-ctx.Done()
//
//	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	if err := srv.Shutdown(shutdownCtx); err != nil {
//	    log.Fatal(err)
//	}
//
// Shutdown drains active connections before returning, giving in-flight
// requests a chance to complete within the provided context's deadline.
package httpserver
