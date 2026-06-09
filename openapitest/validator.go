package openapitest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	libvalidator "github.com/pb33f/libopenapi-validator"
)

// Validator validates HTTP responses against an OpenAPI spec.
type Validator struct {
	v libvalidator.Validator
}

// NewValidator reads the OpenAPI spec at path and returns a Validator.
func NewValidator(path string) (*Validator, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("openapitest: read spec: %w", err)
	}
	return NewValidatorFromSpec(data)
}

// NewValidatorFromSpec parses spec bytes and returns a Validator.
func NewValidatorFromSpec(spec []byte) (*Validator, error) {
	doc, err := libopenapi.NewDocument(spec)
	if err != nil {
		return nil, fmt.Errorf("openapitest: parse spec: %w", err)
	}

	v, errs := libvalidator.NewValidator(doc)
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		return nil, fmt.Errorf("openapitest: build validator: %s", strings.Join(msgs, "; "))
	}

	return &Validator{v: v}, nil
}

// validateResponse reads the body, runs OpenAPI validation, restores the body,
// and returns formatted error messages. Returns nil msgs on success.
func (v *Validator) validateResponse(req *http.Request, resp *http.Response) (raw []byte, msgs []string, err error) {
	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(raw))

	ok, validationErrors := v.v.ValidateHttpResponse(req, resp)
	resp.Body = io.NopCloser(bytes.NewReader(raw))

	if ok {
		return raw, nil, nil
	}

	if len(validationErrors) == 0 {
		return raw, []string{fmt.Sprintf(
			"openapitest: no schema found for %s %s (status %d) — path/operation not in spec",
			req.Method, req.URL.Path, resp.StatusCode,
		)}, nil
	}

	msgs = make([]string, 0, len(validationErrors))
	for _, ve := range validationErrors {
		var b strings.Builder
		fmt.Fprintf(&b, "openapitest: %s\n  reason: %s", ve.Message, ve.Reason)
		for _, se := range ve.SchemaValidationErrors {
			if se.FieldPath != "" {
				fmt.Fprintf(&b, "\n  - %s (at %s)", se.Reason, se.FieldPath)
			} else {
				fmt.Fprintf(&b, "\n  - %s", se.Reason)
			}
		}
		msgs = append(msgs, b.String())
	}
	return raw, msgs, nil
}

// AssertResponse validates resp against the OpenAPI spec for req.
// It calls t.Error for every validation failure, allowing the test to continue.
// The response body is restored after reading so callers can still decode it.
func (v *Validator) AssertResponse(t testing.TB, req *http.Request, resp *http.Response) {
	t.Helper()
	_, msgs, err := v.validateResponse(req, resp)
	if err != nil {
		t.Errorf("openapitest: read response body: %s", err)
		return
	}
	for _, m := range msgs {
		t.Error(m)
	}
}

// RequireResponse is like AssertResponse but calls t.Fatal on the first
// validation failure, stopping the test immediately.
func (v *Validator) RequireResponse(t testing.TB, req *http.Request, resp *http.Response) {
	t.Helper()
	_, msgs, err := v.validateResponse(req, resp)
	if err != nil {
		t.Fatalf("openapitest: read response body: %s", err)
	}
	if len(msgs) > 0 {
		t.Fatal(strings.Join(msgs, "\n"))
	}
}
