package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/brpaz/lib-go/errs"
	"github.com/brpaz/lib-go/logging"
	"github.com/brpaz/lib-go/validator"
)

// defaultStatusMap maps standard errs codes to HTTP status codes.
var defaultStatusMap = map[string]int{
	errs.CodeNotFound:       http.StatusNotFound,
	errs.CodeUnauthorized:   http.StatusUnauthorized,
	errs.CodeForbidden:      http.StatusForbidden,
	errs.CodeInvalidInput:   http.StatusUnprocessableEntity,
	errs.CodeConflict:       http.StatusConflict,
	errs.CodeInternalServer: http.StatusInternalServerError,
	errs.CodeRateLimit:      http.StatusTooManyRequests,
}

// HandleErr maps an error to an HTTP status code and writes a structured JSON response.
// extraCodes extends or overrides the default code→status mapping for app-specific error codes.
// Pass nil if no custom codes are needed.
func HandleErr(w http.ResponseWriter, r *http.Request, err error, extraCodes map[string]int) {
	ctx := r.Context()
	traceID := logging.TraceIDFromContext(ctx)

	var loc Location
	var appErr *errs.Error

	switch {
	case errors.As(err, &appErr):
		// appErr extracted from chain
	case IsJSONDecodeError(err):
		appErr, loc = toJSONAppError(err)
	default:
		appErr = errs.Internal("an unexpected error occurred").WithCause(err)
	}

	writeAppError(w, r, loc, appErr, traceID, extraCodes)
}

// HandleValidationError writes a 422 JSON response for a ValidationError, tagging
// every field error with the given request location (body, query, path, or header).
// The trace_id is populated from the active OTEL span when present.
func HandleValidationError(
	w http.ResponseWriter,
	r *http.Request,
	loc Location,
	ve *validator.ValidationError,
) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)
	logger.Warn(ctx, ve.Message, slog.String("error_code", "invalid_input"))

	traceID := logging.TraceIDFromContext(ctx)

	base := ErrorResponse{
		Code:    "invalid_input",
		Message: ve.Message,
		TraceID: traceID,
	}

	resp := ValidationErrorResponse{
		ErrorResponse: base,
		Errors:        make([]FieldErrorResponse, 0, len(ve.Fields)),
	}
	for _, f := range ve.Fields {
		resp.Errors = append(resp.Errors, FieldErrorResponse{
			Location: string(loc),
			Field:    f.Field,
			Code:     string(f.Code),
			Message:  f.Message,
			Params:   f.Params,
		})
	}

	writeErrorJSON(w, r, http.StatusUnprocessableEntity, resp)
}

func writeAppError(
	w http.ResponseWriter,
	r *http.Request,
	loc Location,
	appErr *errs.Error,
	traceID string,
	extraCodes map[string]int,
) {
	status := resolveStatus(appErr.Code, extraCodes)

	ctx := r.Context()
	logger := logging.FromContext(ctx)

	args := []any{slog.String("error_code", appErr.Code)}
	if cause := appErr.Unwrap(); cause != nil {
		args = append(args, slog.String("error", cause.Error()))
	}
	for k, v := range appErr.Meta {
		args = append(args, slog.Any(k, v))
	}
	if status >= http.StatusInternalServerError {
		logger.Error(ctx, appErr.Message, args...)
	} else {
		logger.Warn(ctx, appErr.Message, args...)
	}

	base := ErrorResponse{
		Code:    appErr.Code,
		Message: appErr.Message,
		TraceID: traceID,
	}

	var payload any
	if len(appErr.Fields) > 0 {
		payload = toValidationErrorResponse(base, loc, appErr.Fields)
	} else {
		payload = base
	}

	writeErrorJSON(w, r, status, payload)
}

func resolveStatus(code string, extraCodes map[string]int) int {
	if status, ok := extraCodes[code]; ok {
		return status
	}
	if status, ok := defaultStatusMap[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

func toValidationErrorResponse(base ErrorResponse, loc Location, fields []errs.FieldError) ValidationErrorResponse {
	ve := ValidationErrorResponse{
		ErrorResponse: base,
		Errors:        make([]FieldErrorResponse, 0, len(fields)),
	}
	for _, f := range fields {
		ve.Errors = append(ve.Errors, FieldErrorResponse{
			Location: string(loc),
			Field:    f.Field,
			Code:     f.Code,
			Message:  f.Message,
			Params:   f.Params,
		})
	}
	return ve
}

func toJSONAppError(err error) (*errs.Error, Location) {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return errs.InvalidInput("invalid request body",
			errs.NewFieldError(
				typeErr.Field,
				string(validator.ErrInvalid),
				fmt.Sprintf("must be %s, got %s", typeErr.Type, typeErr.Value),
			),
		).WithCause(err), LocationBody
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return errs.InvalidInput(
			fmt.Sprintf("malformed JSON at position %d", syntaxErr.Offset),
		).WithCause(err), LocationBody
	}

	return errs.InvalidInput("request body must not be empty").WithCause(err), LocationBody
}

func writeErrorJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger := logging.FromContext(r.Context())
		logger.Error(r.Context(), "failed to encode error response",
			slog.String("error", err.Error()),
		)
	}
}
