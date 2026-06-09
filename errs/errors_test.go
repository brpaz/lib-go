package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/errs"
)

func TestNew(t *testing.T) {
	t.Parallel()

	e := errs.New("MY_CODE", "something went wrong")
	assert.Equal(t, "MY_CODE", e.Code)
	assert.Equal(t, "something went wrong", e.Message)
}

func TestStandardConstructors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  *errs.Error
		code string
	}{
		{"NotFound", errs.NotFound("not found"), errs.CodeNotFound},
		{"Unauthorized", errs.Unauthorized("unauthorized"), errs.CodeUnauthorized},
		{"Forbidden", errs.Forbidden("forbidden"), errs.CodeForbidden},
		{"Conflict", errs.Conflict("conflict"), errs.CodeConflict},
		{"Internal", errs.Internal("internal"), errs.CodeInternalServer},
		{"RateLimitExceeded", errs.RateLimitExceeded("rate limited"), errs.CodeRateLimit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.code, tc.err.Code)
		})
	}
}

func TestInvalidInput_WithFields(t *testing.T) {
	t.Parallel()

	f1 := errs.NewFieldError("email", "required", "email is required")
	f2 := errs.NewFieldError("name", "too_short", "name too short")

	e := errs.InvalidInput("validation failed", f1)
	e.WithFields(f2)

	assert.Equal(t, errs.CodeInvalidInput, e.Code)
	require.Len(t, e.Fields, 2)
	assert.Equal(t, f1, e.Fields[0])
	assert.Equal(t, f2, e.Fields[1])
}

func TestNewFieldErrorWithParams(t *testing.T) {
	t.Parallel()

	params := map[string]any{"min": 3, "max": 50}
	f := errs.NewFieldErrorWithParams("name", "too_short", "name too short", params)

	assert.Equal(t, params, f.Params)
}

func TestWithCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("db connection refused")
	e := errs.Internal("unexpected error").WithCause(cause)

	assert.ErrorIs(t, e, cause)
	assert.Contains(t, e.Error(), cause.Error())
}

func TestWithMeta(t *testing.T) {
	t.Parallel()

	e := errs.NotFound("user not found").
		WithMeta("user_id", "123").
		WithMeta("tenant", "acme")

	assert.Equal(t, "123", e.Meta["user_id"])
	assert.Equal(t, "acme", e.Meta["tenant"])
}

func TestError_String(t *testing.T) {
	t.Parallel()

	t.Run("without cause", func(t *testing.T) {
		t.Parallel()
		e := errs.NotFound("user not found")
		assert.Equal(t, "not_found: user not found", e.Error())
	})

	t.Run("with cause", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("sql: no rows")
		e := errs.NotFound("user not found").WithCause(cause)
		assert.Equal(t, "not_found: user not found: sql: no rows", e.Error())
	})
}

func TestErrorsAs_ThroughWrap(t *testing.T) {
	t.Parallel()

	original := errs.NotFound("user not found")
	wrapped := fmt.Errorf("UserService.GetByID: %w", original)

	var target *errs.Error
	require.True(t, errors.As(wrapped, &target))
	assert.Equal(t, errs.CodeNotFound, target.Code)
}
