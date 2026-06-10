package middleware

import (
	"net/http"
	"slices"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type tracingConfig struct {
	excludedRoutes []string
	excludeFunc    func(r *http.Request) bool
	nameFormatter  func(r *http.Request) string
}

// TracingOption configures the [Tracing] middleware.
type TracingOption func(*tracingConfig)

// WithExcludedRoutes excludes the given exact paths from tracing spans.
// By default, no paths are excluded.
func WithExcludedRoutes(routes ...string) TracingOption {
	return func(c *tracingConfig) { c.excludedRoutes = routes }
}

// WithExcludeFunc excludes requests matching fn from tracing spans. Use it
// for prefix, glob or regex matching beyond the exact paths supported by
// [WithExcludedRoutes].
func WithExcludeFunc(fn func(r *http.Request) bool) TracingOption {
	return func(c *tracingConfig) { c.excludeFunc = fn }
}

// WithSpanNameFormatter sets a function to derive the span name from the
// request. This is useful to use the matched route pattern (e.g. from chi,
// gorilla/mux, etc.) so cardinality stays bounded in the collector.
//
// By default, the span name is "<method> <path>".
func WithSpanNameFormatter(fn func(r *http.Request) string) TracingOption {
	return func(c *tracingConfig) { c.nameFormatter = fn }
}

// Tracing returns an OpenTelemetry HTTP middleware that creates a span per
// request.
func Tracing(opts ...TracingOption) func(http.Handler) http.Handler {
	cfg := &tracingConfig{}
	for _, o := range opts {
		o(cfg)
	}

	otelOpts := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			if cfg.nameFormatter != nil {
				return cfg.nameFormatter(r)
			}
			return r.Method + " " + r.URL.Path
		}),
	}

	if len(cfg.excludedRoutes) > 0 || cfg.excludeFunc != nil {
		otelOpts = append(otelOpts, otelhttp.WithFilter(func(r *http.Request) bool {
			if slices.Contains(cfg.excludedRoutes, r.URL.Path) {
				return false
			}
			return cfg.excludeFunc == nil || !cfg.excludeFunc(r)
		}))
	}

	return otelhttp.NewMiddleware("", otelOpts...)
}
