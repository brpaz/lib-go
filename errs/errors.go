package errs

import "fmt"

// Standard error codes.
const (
	CodeNotFound       = "not_found"
	CodeUnauthorized   = "unauthorized"
	CodeForbidden      = "forbidden"
	CodeInvalidInput   = "invalid_input"
	CodeConflict       = "conflict"
	CodeInternalServer = "internal_error"
	CodeRateLimit      = "rate_limit_exceeded"
)


// FieldError describes a validation failure for a single field.
// It carries no HTTP-specific concepts.
type FieldError struct {
	Field   string
	Code    string
	Message string
	Params  map[string]any
}

// NewFieldError constructs a FieldError without params.
func NewFieldError(field, code, message string) FieldError {
	return FieldError{Field: field, Code: code, Message: message}
}

// NewFieldErrorWithParams constructs a FieldError with extra context params
// (e.g. min_length, max).
func NewFieldErrorWithParams(field, code, message string, params map[string]any) FieldError {
	return FieldError{Field: field, Code: code, Message: message, Params: params}
}

// Error is a structured domain error carrying a machine-readable Code,
// a human-readable Message, optional per-field validation failures, and
// optional internal metadata. It contains no HTTP-specific concepts.
type Error struct {
	Code    string
	Message string
	Fields  []FieldError
	Meta    map[string]any
	cause   error
}

// New constructs an Error with the given code and message.
// Use this for application-specific codes not covered by the standard constructors.
func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// NotFound returns an Error with CodeNotFound.
func NotFound(message string) *Error {
	return &Error{Code: CodeNotFound, Message: message}
}

// Unauthorized returns an Error with CodeUnauthorized.
func Unauthorized(message string) *Error {
	return &Error{Code: CodeUnauthorized, Message: message}
}

// Forbidden returns an Error with CodeForbidden.
func Forbidden(message string) *Error {
	return &Error{Code: CodeForbidden, Message: message}
}

// InvalidInput returns an Error with CodeInvalidInput and optional field errors.
func InvalidInput(message string, fields ...FieldError) *Error {
	return &Error{Code: CodeInvalidInput, Message: message, Fields: fields}
}

// Conflict returns an Error with CodeConflict.
func Conflict(message string) *Error {
	return &Error{Code: CodeConflict, Message: message}
}

// Internal returns an Error with CodeInternalServer.
// Attach the real cause with WithCause so it can be logged without leaking internals.
func Internal(message string) *Error {
	return &Error{Code: CodeInternalServer, Message: message}
}

// RateLimitExceeded returns an Error with CodeRateLimit.
func RateLimitExceeded(message string) *Error {
	return &Error{Code: CodeRateLimit, Message: message}
}

// WithCause attaches an underlying error for internal logging without exposing it to clients.
func (e *Error) WithCause(err error) *Error {
	e.cause = err
	return e
}

// WithMeta attaches a key/value pair for internal logging.
// Meta is never serialised to API responses.
func (e *Error) WithMeta(key string, val any) *Error {
	if e.Meta == nil {
		e.Meta = make(map[string]any)
	}
	e.Meta[key] = val
	return e
}

// WithFields appends field-level validation errors.
func (e *Error) WithFields(fields ...FieldError) *Error {
	e.Fields = append(e.Fields, fields...)
	return e
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause, enabling errors.Is / errors.As traversal.
func (e *Error) Unwrap() error {
	return e.cause
}
