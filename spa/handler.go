package spa

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

// Handler returns an [http.Handler] that serves a single-page application from
// fsys. Static assets are served directly; all other paths are handled by the
// configured not-found handler (default: serve the index file).
func Handler(fsys fs.FS, opts ...Option) http.Handler {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	h := &spaHandler{fs: fsys, opts: o}
	if o.notFoundHandler == nil {
		o.notFoundHandler = http.HandlerFunc(h.serveIndex)
	}
	return h
}

type spaHandler struct {
	fs   fs.FS
	opts *options
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" {
		h.opts.notFoundHandler.ServeHTTP(w, r)
		return
	}

	f, err := h.fs.Open(name)
	if err != nil {
		h.opts.notFoundHandler.ServeHTTP(w, r)
		return
	}
	stat, err := f.Stat()
	f.Close()
	if err != nil || stat.IsDir() {
		h.opts.notFoundHandler.ServeHTTP(w, r)
		return
	}

	if h.opts.staticCacheControl != "" {
		w.Header().Set("Cache-Control", h.opts.staticCacheControl)
	}
	http.FileServerFS(h.fs).ServeHTTP(w, r)
}

func (h *spaHandler) serveIndex(w http.ResponseWriter, _ *http.Request) {
	f, err := h.fs.Open(h.opts.indexFile)
	if err != nil {
		http.Error(w, "index file not found", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if h.opts.envVars != nil {
		data, err := json.Marshal(h.opts.envVars)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		buf = bytes.ReplaceAll(buf, []byte(h.opts.envPlaceholder), data)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", h.opts.indexCacheControl)
	w.Write(buf) //nolint:errcheck
}
