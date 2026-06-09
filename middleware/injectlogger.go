package middleware

import (
	"net/http"

	"github.com/brpaz/lib-go/logging"
)

// InjectLogger returns middleware that stores logger in the request context.
// Downstream handlers retrieve it via [logging.FromContext].
// Use this when you need context logger access without full request logging.
func InjectLogger(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logging.WithContext(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
