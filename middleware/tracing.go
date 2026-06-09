package middleware

import (
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// defaultExcludedRoutes are paths skipped from OTel tracing by default.
var defaultExcludedRoutes = []string{"/", "/healthz", "/metrics"}

type tracingConfig struct {
	excludedRoutes []string
}

// TracingOption configures the [Tracing] middleware.
type TracingOption func(*tracingConfig)

// WithExcludedRoutes overrides the default set of paths excluded from tracing.
func WithExcludedRoutes(routes ...string) TracingOption {
	return func(c *tracingConfig) { c.excludedRoutes = routes }
}

// Tracing returns an OpenTelemetry HTTP middleware that creates a span per
// request. Span names use the matched chi route pattern (e.g. "GET /users/{id}")
// so cardinality stays bounded in the collector.
func Tracing(opts ...TracingOption) func(http.Handler) http.Handler {
	cfg := &tracingConfig{excludedRoutes: defaultExcludedRoutes}
	for _, o := range opts {
		o(cfg)
	}

	excluded := cfg.excludedRoutes
	filter := func(r *http.Request) bool {
		return !slices.Contains(excluded, r.URL.Path)
	}
	namer := func(_ string, r *http.Request) string {
		if pattern := chi.RouteContext(r.Context()).RoutePattern(); pattern != "" {
			return r.Method + " " + pattern
		}
		return r.Method + " " + r.URL.Path
	}

	return otelhttp.NewMiddleware("",
		otelhttp.WithFilter(filter),
		otelhttp.WithSpanNameFormatter(namer),
	)
}
