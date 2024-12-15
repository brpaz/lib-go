package health

import (
	"encoding/json"
	"net/http"
)

// HandlerHealth returns an http.HandlerFunc that handles the health check request
// This function should be registered with your HTTP router to expose the health check endpoint.
//
// Example:
//
//	package main
//
//	import (
//		"net/http"
//		"github.com/brpaz/lib-go/health"
//	)
//
//	func main() {
//		processor := health.New(
//			health.WithName("my-service"),
//			health.WithVersion("1.0.0"),
//			health.WithRevision("abc123"),
//			health.WithChecks(
//				checks.NewStubCheck("stub", true),
//				checks.NewStubCheck("stub2", false),
//			),
//		http.HandleFunc("/health", health.Handler(processor))
//		http.ListenAndServe(":8080", nil)
//	}
func Handler(processor HealthProcessor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		healthResult := processor.Execute(ctx)

		statusCode := http.StatusOK
		if healthResult.Status != StatusPass {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		err := json.NewEncoder(w).Encode(healthResult)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
