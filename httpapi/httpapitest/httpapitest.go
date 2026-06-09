// Package httpapitest provides test helpers for the httpapi package.
package httpapitest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/brpaz/lib-go/httpapi"
)

// AssertFieldError asserts that resp contains a JSON ValidationErrorResponse
// with a field error matching field and code. The response body is restored
// after reading so callers can still decode it.
func AssertFieldError(t testing.TB, resp *http.Response, field, code string) {
	t.Helper()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("httpapitest: read response body: %s", err)
		return
	}
	resp.Body = io.NopCloser(bytes.NewReader(raw))

	var body httpapi.ValidationErrorResponse
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Errorf("httpapitest: response body is not valid JSON: %s", err)
		return
	}

	for _, fe := range body.Errors {
		if fe.Field == field && fe.Code == code {
			return
		}
	}

	t.Errorf("httpapitest: expected field error field=%q code=%q; got %+v", field, code, body.Errors)
}
