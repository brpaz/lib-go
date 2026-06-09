package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/brpaz/lib-go/logging"
)

type requestLoggerConfig struct {
	logRequestBody       bool
	requestBodyMaxBytes  int64
	logResponseBody      bool
	responseBodyMaxBytes int64
}

// RequestLoggerOption configures [RequestLogger].
type RequestLoggerOption func(*requestLoggerConfig)

// WithRequestBodyLogging enables logging of the request body up to maxBytes.
// The body is restored so downstream handlers still receive it.
// Use only in development — request bodies often contain sensitive data.
func WithRequestBodyLogging(maxBytes int64) RequestLoggerOption {
	return func(c *requestLoggerConfig) {
		c.logRequestBody = true
		c.requestBodyMaxBytes = maxBytes
	}
}

// WithResponseBodyLogging enables logging of the response body up to maxBytes.
// Use only in development — response bodies can be large and may contain sensitive data.
func WithResponseBodyLogging(maxBytes int64) RequestLoggerOption {
	return func(c *requestLoggerConfig) {
		c.logResponseBody = true
		c.responseBodyMaxBytes = maxBytes
	}
}

// RequestLogger returns middleware that injects a request-scoped logger into
// the context and logs each request on completion with method, path, status,
// and duration. Downstream handlers retrieve the logger via [logging.FromContext].
func RequestLogger(logger *logging.Logger, opts ...RequestLoggerOption) func(http.Handler) http.Handler {
	cfg := &requestLoggerConfig{}
	for _, o := range opts {
		o(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqLogger := logger.With(
				slog.String("http_method", r.Method),
				slog.String("http_path", r.URL.Path),
				slog.String("http_user_agent", r.UserAgent()),
			)

			ctx := logging.WithContext(r.Context(), reqLogger)
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			var extraAttrs []any

			if cfg.logRequestBody && r.Body != nil {
				body, truncated := limitedRead(r.Body, cfg.requestBodyMaxBytes)
				r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body), r.Body))
				extraAttrs = append(extraAttrs, slog.String("http_request_body", string(body)))
				if truncated {
					extraAttrs = append(extraAttrs, slog.Bool("http_request_body_truncated", true))
				}
			}

			if cfg.logResponseBody {
				rw.captureBody = true
				rw.bodyLimit = cfg.responseBodyMaxBytes
			}

			next.ServeHTTP(rw, r.WithContext(ctx))

			completionAttrs := append([]any{
				slog.Int("http_status", rw.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			}, extraAttrs...)

			if cfg.logResponseBody {
				completionAttrs = append(completionAttrs,
					slog.String("http_response_body", string(rw.body)),
				)
				if rw.bodyTruncated {
					completionAttrs = append(completionAttrs,
						slog.Bool("http_response_body_truncated", true),
					)
				}
			}

			reqLogger.Info(ctx, "request completed", completionAttrs...)
		})
	}
}

// limitedRead reads up to maxBytes from r. Returns the data and whether it was truncated.
func limitedRead(r io.Reader, maxBytes int64) ([]byte, bool) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil || int64(len(data)) <= maxBytes {
		return data, false
	}
	return data[:maxBytes], true
}

type responseWriter struct {
	http.ResponseWriter
	status        int
	captureBody   bool
	bodyLimit     int64
	body          []byte
	bodyTruncated bool
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.captureBody {
		remaining := rw.bodyLimit - int64(len(rw.body))
		if remaining > 0 {
			take := int64(len(b))
			if take > remaining {
				take = remaining
				rw.bodyTruncated = true
			}
			rw.body = append(rw.body, b[:take]...)
		} else if len(b) > 0 {
			rw.bodyTruncated = true
		}
	}
	return rw.ResponseWriter.Write(b)
}
