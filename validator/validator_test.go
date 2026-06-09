package validator_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/validator"
)

func TestValidator_Result(t *testing.T) {
	t.Parallel()

	t.Run("no errors returns nil", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("email", "user@example.com", validator.Required(), validator.Email())
		assert.Nil(t, v.Result())
	})

	t.Run("returns ValidationError with all field errors", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("email", "", validator.Required())
		v.Field("password", "", validator.Required())

		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, "validation failed", result.Message)
		assert.Len(t, result.Fields, 2)
	})

	t.Run("stops at first failure per field", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("email", "", validator.Required(), validator.Email())

		result := v.Result()
		require.NotNil(t, result)
		assert.Len(t, result.Fields, 1)
		assert.Equal(t, validator.ErrRequired, result.Fields[0].Code)
	})

	t.Run("nil value treated as empty", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("name", nil, validator.Required())

		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, validator.ErrRequired, result.Fields[0].Code)
	})

	t.Run("nil *string treated as empty", func(t *testing.T) {
		t.Parallel()

		var name *string
		v := validator.New()
		v.Field("name", name, validator.Required())

		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, validator.ErrRequired, result.Fields[0].Code)
	})

	t.Run("non-nil *string passes through", func(t *testing.T) {
		t.Parallel()

		name := "alice"
		v := validator.New()
		v.Field("name", &name, validator.Required())

		assert.Nil(t, v.Result())
	})

	t.Run("non-ruleError captured with ErrInvalid", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("custom", "value", func(any) error {
			return errors.New("external validation failed")
		})

		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, validator.ErrInvalid, result.Fields[0].Code)
		assert.Equal(t, "external validation failed", result.Fields[0].Message)
	})
}

func TestRequired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "non-empty string", value: "hello", wantErr: false},
		{name: "empty string", value: "", wantErr: true},
		{name: "whitespace only", value: "   ", wantErr: true},
		{name: "tab only", value: "\t", wantErr: true},
		{name: "nil", value: nil, wantErr: true},
		{name: "non-nil int is considered present", value: 0, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.Required()(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "valid email", value: "user@example.com", wantErr: false},
		{name: "valid with subdomain", value: "user@mail.example.com", wantErr: false},
		{name: "missing @", value: "userexample.com", wantErr: true},
		{name: "missing domain", value: "user@", wantErr: true},
		{name: "plain text", value: "notanemail", wantErr: true},
		{name: "empty string", value: "", wantErr: true},
		{name: "display-name format rejected", value: "John <user@example.com>", wantErr: true},
		{name: "angle-bracket format rejected", value: "<user@example.com>", wantErr: true},
		{name: "wrong type", value: 42, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.Email()(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		min     int
		wantErr bool
	}{
		{name: "at minimum", value: "abc", min: 3, wantErr: false},
		{name: "above minimum", value: "abcd", min: 3, wantErr: false},
		{name: "below minimum", value: "ab", min: 3, wantErr: true},
		{name: "empty below minimum", value: "", min: 1, wantErr: true},
		{name: "multibyte runes counted correctly", value: "héllo", min: 5, wantErr: false},
		{name: "wrong type", value: 42, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.MinLength(tt.min)(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("params carry min", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("name", "ab", validator.MinLength(5))
		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, 5, result.Fields[0].Params["min"])
	})
}

func TestMaxLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		max     int
		wantErr bool
	}{
		{name: "at maximum", value: "abc", max: 3, wantErr: false},
		{name: "below maximum", value: "ab", max: 3, wantErr: false},
		{name: "above maximum", value: "abcd", max: 3, wantErr: true},
		{name: "empty always valid", value: "", max: 5, wantErr: false},
		{name: "multibyte runes counted correctly", value: "héllo", max: 4, wantErr: true},
		{name: "wrong type", value: 42, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.MaxLength(tt.max)(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("params carry max", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("name", "toolong", validator.MaxLength(3))
		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, 3, result.Fields[0].Params["max"])
	})
}

func TestHasSpecialChar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "has exclamation", value: "abc!def", wantErr: false},
		{name: "has at-sign", value: "abc@def", wantErr: false},
		{name: "has space", value: "abc def", wantErr: false},
		{name: "has hyphen", value: "abc-def", wantErr: false},
		{name: "letters only", value: "abcdefgh", wantErr: true},
		{name: "letters and digits only", value: "abc123", wantErr: true},
		{name: "empty", value: "", wantErr: true},
		{name: "wrong type", value: 42, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.HasSpecialChar()(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		wantErr  bool
		wantCode validator.ErrCode
	}{
		{name: "above threshold", value: 5, wantErr: false},
		{name: "int64 above threshold", value: int64(5), wantErr: false},
		{name: "equal to threshold", value: 3, wantErr: true, wantCode: validator.ErrMin},
		{name: "below threshold", value: 1, wantErr: true, wantCode: validator.ErrMin},
		{name: "string type", value: "5", wantErr: true, wantCode: validator.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := validator.New()
			v.Field("n", tt.value, validator.Gt(3))
			result := v.Result()

			if tt.wantErr {
				require.NotNil(t, result)
				if tt.wantCode != "" {
					assert.Equal(t, tt.wantCode, result.Fields[0].Code)
				}
			} else {
				require.Nil(t, result)
			}
		})
	}

	t.Run("params carry gt", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		v.Field("n", 1, validator.Gt(3))
		result := v.Result()
		require.NotNil(t, result)
		assert.Equal(t, float64(3), result.Fields[0].Params["gt"])
	})
}

func TestGte(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		wantErr  bool
		wantCode validator.ErrCode
	}{
		{name: "above threshold", value: 5, wantErr: false},
		{name: "equal to threshold", value: 3, wantErr: false},
		{name: "below threshold", value: 1, wantErr: true, wantCode: validator.ErrMin},
		{name: "string type", value: "5", wantErr: true, wantCode: validator.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := validator.New()
			v.Field("n", tt.value, validator.Gte(3))
			result := v.Result()

			if tt.wantErr {
				require.NotNil(t, result)
				if tt.wantCode != "" {
					assert.Equal(t, tt.wantCode, result.Fields[0].Code)
				}
			} else {
				require.Nil(t, result)
			}
		})
	}
}

func TestLt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		wantErr  bool
		wantCode validator.ErrCode
	}{
		{name: "below threshold", value: 1, wantErr: false},
		{name: "equal to threshold", value: 3, wantErr: true, wantCode: validator.ErrMax},
		{name: "above threshold", value: 5, wantErr: true, wantCode: validator.ErrMax},
		{name: "string type", value: "1", wantErr: true, wantCode: validator.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := validator.New()
			v.Field("n", tt.value, validator.Lt(3))
			result := v.Result()

			if tt.wantErr {
				require.NotNil(t, result)
				if tt.wantCode != "" {
					assert.Equal(t, tt.wantCode, result.Fields[0].Code)
				}
			} else {
				require.Nil(t, result)
			}
		})
	}
}

func TestLte(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    any
		wantErr  bool
		wantCode validator.ErrCode
	}{
		{name: "below threshold", value: 1, wantErr: false},
		{name: "equal to threshold", value: 3, wantErr: false},
		{name: "above threshold", value: 5, wantErr: true, wantCode: validator.ErrMax},
		{name: "string type", value: "1", wantErr: true, wantCode: validator.ErrInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := validator.New()
			v.Field("n", tt.value, validator.Lte(3))
			result := v.Result()

			if tt.wantErr {
				require.NotNil(t, result)
				if tt.wantCode != "" {
					assert.Equal(t, tt.wantCode, result.Fields[0].Code)
				}
			} else {
				require.Nil(t, result)
			}
		})
	}
}
