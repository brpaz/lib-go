package middleware

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/brpaz/lib-go/log"
)

// LoggableResponseWriter is a wrapper around http.ResponseWriter that captures the status code, so that we can retrieve it in the logging middleware.
type LoggableResponseWriter struct {
	http.ResponseWriter
	StatusCode   int
	ResponseSize int
	Written      bool
}

// WriteHeader writes the status code to the response.
func (lrw *LoggableResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write writes the data to the response and tracks the size of the written data.
func (lrw *LoggableResponseWriter) Write(p []byte) (int, error) {
	if !lrw.Written {
		lrw.Written = true
	}
	n, err := lrw.ResponseWriter.Write(p)
	lrw.ResponseSize += n
	return n, err
}

// RequestLoggerConfig holds configuration options for logging middleware.
type RequestLoggerConfig struct {
	LogMethod             bool
	LogPath               bool
	LogStatus             bool
	LogDuration           bool
	LogRemoteAddr         bool
	LogUserAgent          bool
	LogRequestSize        bool
	LogResponseSize       bool
	LogRequestQueryParams bool
	LogRequestHeaders     bool
	LogRequestBody        bool
	LogResponseBody       bool
}

// defaultRequestLoggerConfig provides the default logging configuration.
var defaultRequestLoggerConfig = RequestLoggerConfig{
	LogMethod:             true,
	LogPath:               true,
	LogStatus:             true,
	LogDuration:           true,
	LogRemoteAddr:         true,
	LogUserAgent:          true,
	LogRequestSize:        true,
	LogRequestQueryParams: false,
	LogRequestHeaders:     false,
	LogRequestBody:        false,
	LogResponseSize:       true,
}

func DefaultRequestLoggerConfig() RequestLoggerConfig {
	return defaultRequestLoggerConfig
}

// RequestLogger is a middleware that logs the http request.
//
// Usage:
//
//	package main
//	import (
//		"net/http"
//	 	"github.com/brpaz/lib-go/log"
//		"github.com/brpaz/lib-go/log/middleware"
//	)
//
//	func main() {
//		logger, _ := log.New(log.WithAdapter("zap"))
//		http.Handle("/", middleware.RequestLogger(logger, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		        w.Write([]byte("Hello, World!"))
//		}))
//		http.ListenAndServe(":8080", nil)
//	}
func RequestLogger(logger log.Logger, config *RequestLoggerConfig) func(http.Handler) http.Handler {
	// Use default config if none provided
	if config == nil {
		config = &defaultRequestLoggerConfig
	}

	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, r *http.Request) {
			t1 := time.Now()
			lrw := &LoggableResponseWriter{rw, http.StatusOK, 0, false}
			defer func() {
				logFields := buildLogFieldsFromConfig(r, lrw, config, t1)
				// Log the request details
				logger.Info(r.Context(), "incoming request", logFields...)
			}()
			next.ServeHTTP(lrw, r)
		}

		return http.HandlerFunc(fn)
	}
}

func buildLogFieldsFromConfig(r *http.Request, lrw *LoggableResponseWriter, config *RequestLoggerConfig, reqStartTime time.Time) []log.Field {
	fields := []log.Field{}

	// Conditionally log based on the configuration
	if config.LogMethod {
		fields = append(fields, log.String("method", r.Method))
	}
	if config.LogPath {
		fields = append(fields, log.String("path", r.URL.Path))
	}
	if config.LogStatus {
		fields = append(fields, log.Int("status", lrw.StatusCode))
	}
	if config.LogDuration {
		fields = append(fields, log.String("duration", time.Since(reqStartTime).String()))
	}
	if config.LogRequestSize {
		fields = append(fields, log.Int("request_size", int(r.ContentLength)))
	}
	if config.LogResponseSize {
		fields = append(fields, log.Int("response_size", lrw.ResponseSize))
	}

	if config.LogRequestQueryParams {
		fields = append(fields, log.String("query", r.URL.RawQuery))
	}

	if config.LogRemoteAddr {
		fields = append(fields, log.String("remote_addr", r.RemoteAddr))
	}

	if config.LogUserAgent {
		fields = append(fields, log.String("user_agent", r.UserAgent()))
	}

	if config.LogRequestBody {
		body, _ := io.ReadAll(r.Body)

		fields = append(fields, log.String("request_body", string(body)))
	}

	if config.LogRequestHeaders {
		for key, values := range r.Header {
			for _, value := range values {
				fields = append(fields, log.String("header_"+strings.ToLower(key), value))
			}
		}
	}
	return fields
}
