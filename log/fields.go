package log

import (
	"go.uber.org/zap"
)

// Field defines a structured log attribute that allows to add additional information to a log entry in a type-safe way.
// To avoid reinventing the wheel, this is just a wrapper around slog.Attr, from the slog package. We can changhe the underlying implementation in the future if needed, but the interface for the rest of the application will remain the same.
type Field = zap.Field

var (
	Bool     = zap.Bool
	Int      = zap.Int
	Int64    = zap.Int64
	Float64  = zap.Float64
	String   = zap.String
	Duration = zap.Duration
	Float32  = zap.Float32
	Float64p = zap.Float64p
	Error    = zap.Error
	Any      = zap.Any
)
