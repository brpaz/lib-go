package apidoc

import (
	"fmt"
	"net/http"
)

const scalarHTMLTemplate = `<!doctype html>
<html>
  <head>
    <title>API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="./openapi.yml"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

// Handler serves OpenAPI documentation. It implements [http.Handler] and is
// intended to be mounted at a path prefix (e.g. chi's router.Mount or
// http.StripPrefix + http.Handle).
//
//	GET /             → Scalar UI (requires [WithScalarUI])
//	GET /openapi.yml  → raw YAML spec
type Handler struct {
	spec    []byte
	serveUI bool
}

// Option configures a [Handler].
type Option func(*Handler)

// WithSpec sets the raw OpenAPI spec bytes to serve.
func WithSpec(spec []byte) Option {
	return func(h *Handler) { h.spec = spec }
}

// WithScalarUI enables the Scalar UI route at the handler root.
func WithScalarUI() Option {
	return func(h *Handler) { h.serveUI = true }
}

// New creates a new [Handler] and validates required options.
func New(opts ...Option) (*Handler, error) {
	h := &Handler{}
	for _, opt := range opts {
		opt(h)
	}
	if len(h.spec) == 0 {
		return nil, fmt.Errorf("apidoc: OpenAPI spec is required")
	}
	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "", "/":
		if !h.serveUI {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, scalarHTMLTemplate)
	case "/openapi.yml":
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(h.spec)
	default:
		http.NotFound(w, r)
	}
}
