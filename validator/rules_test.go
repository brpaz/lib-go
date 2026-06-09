package validator_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/validator"
)

// --- string rules ---

func TestUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "valid v4 UUID", value: "550e8400-e29b-41d4-a716-446655440000", wantErr: false},
		{name: "valid v1 UUID", value: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", wantErr: false},
		{name: "uppercase UUID", value: "550E8400-E29B-41D4-A716-446655440000", wantErr: false},
		{name: "empty string", value: "", wantErr: true},
		{name: "not a UUID", value: "not-a-uuid", wantErr: true},
		{name: "missing segment", value: "550e8400-e29b-41d4-a716", wantErr: true},
		{name: "plain string", value: "hello", wantErr: true},
		{name: "wrong type", value: 42, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.UUID()(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "valid http URL", value: "http://example.com", wantErr: false},
		{name: "valid https URL", value: "https://example.com/path?q=1", wantErr: false},
		{name: "valid URL with port", value: "http://localhost:8080", wantErr: false},
		{name: "empty string", value: "", wantErr: true},
		{name: "no scheme", value: "example.com", wantErr: true},
		{name: "scheme only", value: "https://", wantErr: true},
		{name: "plain string", value: "not a url", wantErr: true},
		{name: "wrong type", value: 42, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.URL()(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	t.Parallel()

	alphanumeric := regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	t.Run("default message", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			value   any
			wantErr bool
		}{
			{name: "matches pattern", value: "abc123", wantErr: false},
			{name: "uppercase matches", value: "ABC", wantErr: false},
			{name: "contains special char", value: "abc!", wantErr: true},
			{name: "empty string", value: "", wantErr: true},
			{name: "space in string", value: "ab cd", wantErr: true},
			{name: "wrong type", value: 42, wantErr: true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				err := validator.Matches(alphanumeric)(tt.value)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("custom message", func(t *testing.T) {
		t.Parallel()

		err := validator.Matches(alphanumeric, "must be alphanumeric")("abc!")
		require.Error(t, err)
		assert.Equal(t, "must be alphanumeric", err.Error())
	})

	t.Run("pattern in params", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("slug", "abc!", validator.Matches(alphanumeric))
		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, alphanumeric.String(), result.Fields[0].Params["pattern"])
	})
}

// --- numeric rules ---

func TestBetween(t *testing.T) {
	t.Parallel()

	t.Run("float64 values", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			value    any
			wantErr  bool
			wantCode validator.ErrCode
		}{
			{name: "within range", value: float64(5), wantErr: false},
			{name: "at min boundary", value: float64(1), wantErr: false},
			{name: "at max boundary", value: float64(10), wantErr: false},
			{name: "below min", value: float64(0), wantErr: true, wantCode: validator.ErrMin},
			{name: "above max", value: float64(11), wantErr: true, wantCode: validator.ErrMax},
			{name: "type mismatch", value: "5", wantErr: true, wantCode: validator.ErrInvalid},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				v := validator.New()
				v.Field("n", tt.value, validator.Between(float64(1), float64(10)))
				result := v.Result()

				if tt.wantErr {
					require.NotNil(t, result)
					if tt.wantCode != "" {
						require.Equal(t, tt.wantCode, result.Fields[0].Code)
					}
				} else {
					require.Nil(t, result)
				}
			})
		}
	})

	t.Run("int values", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validator.Between(1, 10)(5))
		require.Error(t, validator.Between(1, 10)(0))
		require.Error(t, validator.Between(1, 10)(11))
	})

	t.Run("params carry min and max", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("n", float64(0), validator.Between(float64(1), float64(10)))
		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, float64(1), result.Fields[0].Params["min"])
		assert.Equal(t, float64(10), result.Fields[0].Params["max"])
	})
}

func TestPositive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "int above zero", value: 1, wantErr: false},
		{name: "int64 above zero", value: int64(1), wantErr: false},
		{name: "float above zero", value: 0.1, wantErr: false},
		{name: "zero fails", value: 0, wantErr: true},
		{name: "negative fails", value: -1, wantErr: true},
		{name: "wrong type", value: "1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.Positive()(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNonNegative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "zero passes", value: 0, wantErr: false},
		{name: "int above zero", value: 1, wantErr: false},
		{name: "int64 above zero", value: int64(1), wantErr: false},
		{name: "negative fails", value: -1, wantErr: true},
		{name: "wrong type", value: "0", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.NonNegative()(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- generic rules ---

func TestOneOf(t *testing.T) {
	t.Parallel()

	t.Run("string values", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name    string
			value   any
			wantErr bool
		}{
			{name: "allowed value", value: "admin", wantErr: false},
			{name: "another allowed value", value: "user", wantErr: false},
			{name: "not allowed", value: "superuser", wantErr: true},
			{name: "empty string not allowed", value: "", wantErr: true},
			{name: "wrong type", value: 1, wantErr: true},
		}

		rule := validator.OneOf("admin", "user", "moderator")

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				err := rule(tt.value)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("int values", func(t *testing.T) {
		t.Parallel()

		rule := validator.OneOf(1, 2, 3)

		require.NoError(t, rule(1))
		require.NoError(t, rule(2))
		require.Error(t, rule(4))
		require.Error(t, rule("1"))
	})
}

func TestOptional(t *testing.T) {
	t.Parallel()

	t.Run("skips rules for empty string", func(t *testing.T) {
		t.Parallel()

		err := validator.Optional(validator.Email())("")
		require.NoError(t, err)
	})

	t.Run("skips rules for whitespace-only string", func(t *testing.T) {
		t.Parallel()

		err := validator.Optional(validator.Email())("   ")
		require.NoError(t, err)
	})

	t.Run("skips rules for nil", func(t *testing.T) {
		t.Parallel()

		err := validator.Optional(validator.Email())(nil)
		require.NoError(t, err)
	})

	t.Run("skips rules for nil *string", func(t *testing.T) {
		t.Parallel()

		var s *string
		err := validator.Optional(validator.Email())(s)
		require.NoError(t, err)
	})

	t.Run("applies rules for non-empty value", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validator.Optional(validator.Email())("user@example.com"))
		require.Error(t, validator.Optional(validator.Email())("notanemail"))
	})

	t.Run("applies multiple rules in order", func(t *testing.T) {
		t.Parallel()

		rule := validator.Optional(validator.MinLength(3), validator.MaxLength(10))

		require.NoError(t, rule("hello"))
		require.Error(t, rule("hi"))
		require.Error(t, rule("toolongstring"))
	})
}
