package httpapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/httpapi"
)

func TestDecodeJSON(t *testing.T) {
	t.Parallel()

	t.Run("decodes valid body", func(t *testing.T) {
		t.Parallel()
		body := bytes.NewBufferString(`{"name":"alice","age":30}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)
		got, err := httpapi.DecodeJSON[testPayload](r)
		require.NoError(t, err)
		assert.Equal(t, testPayload{Name: "alice", Age: 30}, got)
	})

	t.Run("error on malformed JSON", func(t *testing.T) {
		t.Parallel()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad`))
		_, err := httpapi.DecodeJSON[testPayload](r)
		require.Error(t, err)
	})

	t.Run("error on empty body", func(t *testing.T) {
		t.Parallel()
		r := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
		_, err := httpapi.DecodeJSON[testPayload](r)
		require.Error(t, err)
	})

	t.Run("error on type mismatch", func(t *testing.T) {
		t.Parallel()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"age":"not-a-number"}`))
		_, err := httpapi.DecodeJSON[testPayload](r)
		require.Error(t, err)
	})
}
