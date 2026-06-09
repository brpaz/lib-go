package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// DecodeJSON decodes the JSON request body into T.
// Returns an error if the body is missing, malformed, or cannot be decoded.
func DecodeJSON[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode request body: %w", err)
	}
	return v, nil
}

// IsJSONDecodeError reports whether err originated from JSON decoding:
// syntax errors, type mismatches, or an empty/truncated body.
func IsJSONDecodeError(err error) bool {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	return errors.As(err, &syntaxErr) ||
		errors.As(err, &typeErr) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}
