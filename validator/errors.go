package validator

import "fmt"

// ErrCode is a machine-readable validation error code.
type ErrCode string

// Field-level validation codes.
const (
	ErrRequired  ErrCode = "required"
	ErrInvalid   ErrCode = "invalid"
	ErrMinLength ErrCode = "min_length"
	ErrMaxLength ErrCode = "max_length"
	ErrMin       ErrCode = "min"
	ErrMax       ErrCode = "max"
	ErrDuplicate ErrCode = "duplicate"
)

// RuleError is returned by a Rule when validation fails.
// Callers that invoke rules directly can use errors.As to retrieve the code and params.
type RuleError struct {
	Code    ErrCode
	Message string
	Params  map[string]any
}

// Error implements the error interface.
func (e *RuleError) Error() string { return e.Message }

// FieldError describes a validation failure for a single field.
type FieldError struct {
	Field   string
	Code    ErrCode
	Message string
	Params  map[string]any
}

// Error implements the error interface.
func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationError is returned by [Validator.Result] when one or more fields failed validation.
type ValidationError struct {
	Message string
	Fields  []FieldError
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s (%d field errors)", e.Message, len(e.Fields))
}
