package spa

import "net/http"

const defaultIndexCacheControl = "no-cache, no-store, must-revalidate"

// Option configures a SPA handler.
type Option func(*options)

type options struct {
	indexFile           string
	envPlaceholder      string
	envVars             map[string]any
	notFoundHandler     http.Handler
	indexCacheControl   string
	staticCacheControl  string
}

func defaultOptions() *options {
	return &options{
		indexFile:         "index.html",
		indexCacheControl: defaultIndexCacheControl,
	}
}

// WithIndexFile sets the path within the FS used as the SPA entry point.
// Defaults to "index.html".
func WithIndexFile(name string) Option {
	return func(o *options) {
		o.indexFile = name
	}
}

// WithEnvVars injects runtime environment variables into the index file.
// Every occurrence of placeholder in the index file is replaced with a
// JSON-encoded representation of vars before the response is written.
func WithEnvVars(placeholder string, vars map[string]any) Option {
	return func(o *options) {
		o.envPlaceholder = placeholder
		o.envVars = vars
	}
}

// WithNotFoundHandler sets the handler called when a requested path does not
// exist as a static asset. Defaults to serving the index file (SPA fallback).
func WithNotFoundHandler(h http.Handler) Option {
	return func(o *options) {
		o.notFoundHandler = h
	}
}

// WithIndexCacheControl sets the Cache-Control header on index file responses.
// Defaults to "no-cache, no-store, must-revalidate".
func WithIndexCacheControl(value string) Option {
	return func(o *options) {
		o.indexCacheControl = value
	}
}

// WithStaticCacheControl sets the Cache-Control header on static asset responses.
// For SPAs with content-hashed filenames, "public, max-age=31536000, immutable" is typical.
// If unset, no Cache-Control header is added (http.FileServerFS handles ETags/Last-Modified).
func WithStaticCacheControl(value string) Option {
	return func(o *options) {
		o.staticCacheControl = value
	}
}
