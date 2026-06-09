// Package spa provides an [http.Handler] that serves single-page applications.
//
// Static assets are served directly from an [io/fs.FS]; all other request
// paths fall back to the index file so that client-side routing works without
// server-side route configuration.
//
// # Basic usage
//
//	//go:embed dist
//	var distFS embed.FS
//
//	sub, _ := fs.Sub(distFS, "dist")
//	handler := spa.Handler(sub)
//
// # Runtime environment injection
//
// If your index file contains a placeholder string, [WithEnvVars] replaces it
// with a JSON object at request time:
//
//	handler := spa.Handler(sub,
//	    spa.WithEnvVars("__FRONTEND_ENV__", map[string]any{
//	        "API_URL": os.Getenv("API_URL"),
//	    }),
//	)
//
// # Custom options
//
//	handler := spa.Handler(sub,
//	    spa.WithIndexFile("shell.html"),
//	    spa.WithIndexCacheControl("no-cache"),
//	    spa.WithNotFoundHandler(http.NotFoundHandler()),
//	)
package spa
