// Package apidoc provides an HTTP handler for serving OpenAPI documentation.
//
// It exposes two routes:
//
//   - GET /openapi.yml — raw OpenAPI spec (always enabled)
//   - GET /           — Scalar UI (opt-in via [WithScalarUI])
//
// # Basic usage
//
//	spec, _ := os.ReadFile("openapi.yml")
//
//	h, err := apidoc.New(
//	    apidoc.WithSpec(spec),
//	    apidoc.WithScalarUI(),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Mount under a prefix with chi:
//	r.Mount("/docs", h)
//
//	// Or with stdlib:
//	http.Handle("/docs/", http.StripPrefix("/docs", h))
package apidoc
