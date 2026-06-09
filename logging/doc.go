// Package logging provides a structured logger built on [log/slog] with
// support for OpenTelemetry trace/span injection, runtime level control,
// and a middleware chain for application-specific attribute injection.
//
// # Construction
//
// Create a logger once at application startup using functional options:
//
//	logger, err := logging.NewLogger(
//		logging.WithEnvironment("production"),
//		logging.WithFormat(logging.FormatJSON),
//		logging.WithVersion("1.8.0"),
//		logging.WithRevision("a1b2c3d4"),
//		logging.WithLevel("info"),
//		logging.WithMiddleware(logging.OtelMiddleware()),
//	)
//
// The returned [Logger] exposes [Logger.SetLevel] and [Logger.GetLevel] to
// change the active log level at runtime without restarting the process:
//
//	logger.SetLevel(slog.LevelDebug)
//
// # Logging
//
// Every method accepts a [context.Context] so middlewares can extract
// request-scoped values (trace IDs, user IDs, etc.) automatically:
//
//	logger.Info(ctx, "user authenticated", slog.String("user_id", id))
//	logger.Error(ctx, "payment failed", slog.Any("error", err))
//
// # Middleware
//
// Use [WithMiddleware] to inject application-specific attributes on every log
// call. A middleware receives the [slog.Handler] it wraps and the context on
// each Handle call, making it ideal for pulling values out of ctx:
//
//	func UserIDMiddleware(next slog.Handler) logging.Middleware {
//		return func(next slog.Handler) slog.Handler {
//			return &userIDHandler{Handler: next}
//		}
//	}
//
// Pass one or more middlewares at construction time:
//
//	logging.NewLogger(
//		logging.WithMiddleware(
//			logging.OtelMiddleware(),
//			UserIDMiddleware,
//		),
//	)
//
// # Extending with static attributes
//
// Use [Logger.With] to derive a child logger pre-populated with extra fields.
// Call it at the request boundary and pass the enriched logger down:
//
//	reqLogger := logger.With(slog.String("request_id", id))
//	reqLogger.Info(ctx, "request started")
//
// # Context propagation
//
// Store and retrieve a logger from a [context.Context]:
//
//	ctx = logging.WithContext(ctx, logger)
//	logging.FromContext(ctx).Info(ctx, "retrieved from context")
//
// [FromContext] returns a no-op logger when none is stored, so callers never
// need to nil-check.
package logging
