package validator

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// --- string rules ---

// Required fails when value is nil, empty, or whitespace-only.
// Non-nil, non-string values are considered present and pass.
func Required() Rule {
	return func(value any) error {
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return &RuleError{Code: ErrRequired, Message: "this field is required"}
			}
		case *string:
			if v == nil || strings.TrimSpace(*v) == "" {
				return &RuleError{Code: ErrRequired, Message: "this field is required"}
			}
		default:
			if value == nil {
				return &RuleError{Code: ErrRequired, Message: "this field is required"}
			}
		}

		return nil
	}
}

// MinLength fails when the UTF-8 rune count of value is less than n.
func MinLength(n int) Rule {
	return func(value any) error {
		s, ok := asString(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a string value"}
		}

		if len([]rune(s)) < n {
			return &RuleError{
				Code:    ErrMinLength,
				Message: fmt.Sprintf("must be at least %d characters", n),
				Params:  map[string]any{"min": n},
			}
		}

		return nil
	}
}

// MaxLength fails when the UTF-8 rune count of value exceeds n.
func MaxLength(n int) Rule {
	return func(value any) error {
		s, ok := asString(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a string value"}
		}

		if len([]rune(s)) > n {
			return &RuleError{
				Code:    ErrMaxLength,
				Message: fmt.Sprintf("must be at most %d characters", n),
				Params:  map[string]any{"max": n},
			}
		}

		return nil
	}
}

// Email fails when the value is not a valid email address.
// Display-name format ("John <user@example.com>") is rejected; only bare addresses pass.
func Email() Rule {
	return func(value any) error {
		s, ok := asString(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a string value"}
		}

		addr, err := mail.ParseAddress(s)
		if err != nil || addr.Address != strings.TrimSpace(s) {
			return &RuleError{Code: ErrInvalid, Message: "invalid email format"}
		}

		return nil
	}
}

// UUID fails when the value is not a valid RFC 4122 UUID.
func UUID() Rule {
	return func(value any) error {
		s, ok := asString(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a string value"}
		}

		if _, err := uuid.Parse(s); err != nil {
			return &RuleError{Code: ErrInvalid, Message: "invalid UUID format"}
		}

		return nil
	}
}

// URL fails when the value is not a valid absolute URL (scheme and host required).
func URL() Rule {
	return func(value any) error {
		s, ok := asString(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a string value"}
		}

		u, err := url.Parse(s)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return &RuleError{Code: ErrInvalid, Message: "invalid URL format"}
		}

		return nil
	}
}

// Matches fails when the value does not match re.
// Compile the regexp once at package or init level, not inside a handler.
// An optional message overrides the default "invalid format" error text.
func Matches(re *regexp.Regexp, message ...string) Rule {
	msg := "invalid format"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	return func(value any) error {
		s, ok := asString(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a string value"}
		}

		if !re.MatchString(s) {
			return &RuleError{
				Code:    ErrInvalid,
				Message: msg,
				Params:  map[string]any{"pattern": re.String()},
			}
		}

		return nil
	}
}

// HasSpecialChar fails when the value contains no non-alphanumeric character.
func HasSpecialChar() Rule {
	return func(value any) error {
		s, ok := asString(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a string value"}
		}

		if !containsSpecial(s) {
			return &RuleError{Code: ErrInvalid, Message: "must contain at least one special character"}
		}

		return nil
	}
}

// --- numeric rules ---

// Gt fails when the numeric value is not greater than n.
func Gt(n float64) Rule {
	return func(value any) error {
		v, ok := toFloat64(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a numeric value"}
		}

		if v <= n {
			return &RuleError{
				Code:    ErrMin,
				Message: fmt.Sprintf("must be greater than %v", n),
				Params:  map[string]any{"gt": n},
			}
		}

		return nil
	}
}

// Gte fails when the numeric value is less than n.
func Gte(n float64) Rule {
	return func(value any) error {
		v, ok := toFloat64(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a numeric value"}
		}

		if v < n {
			return &RuleError{
				Code:    ErrMin,
				Message: fmt.Sprintf("must be greater than or equal to %v", n),
				Params:  map[string]any{"gte": n},
			}
		}

		return nil
	}
}

// Lt fails when the numeric value is not less than n.
func Lt(n float64) Rule {
	return func(value any) error {
		v, ok := toFloat64(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a numeric value"}
		}

		if v >= n {
			return &RuleError{
				Code:    ErrMax,
				Message: fmt.Sprintf("must be less than %v", n),
				Params:  map[string]any{"lt": n},
			}
		}

		return nil
	}
}

// Lte fails when the numeric value exceeds n.
func Lte(n float64) Rule {
	return func(value any) error {
		v, ok := toFloat64(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a numeric value"}
		}

		if v > n {
			return &RuleError{
				Code:    ErrMax,
				Message: fmt.Sprintf("must be less than or equal to %v", n),
				Params:  map[string]any{"lte": n},
			}
		}

		return nil
	}
}

// Between fails when value is outside [min, max] (inclusive).
// Returns ErrMin when below min, ErrMax when above max.
func Between(min, max float64) Rule {
	return func(value any) error {
		v, ok := toFloat64(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a numeric value"}
		}

		if v < min {
			return &RuleError{
				Code:    ErrMin,
				Message: fmt.Sprintf("must be at least %v", min),
				Params:  map[string]any{"min": min, "max": max},
			}
		}

		if v > max {
			return &RuleError{
				Code:    ErrMax,
				Message: fmt.Sprintf("must be at most %v", max),
				Params:  map[string]any{"min": min, "max": max},
			}
		}

		return nil
	}
}

// Positive fails when value is not greater than 0.
func Positive() Rule {
	return func(value any) error {
		v, ok := toFloat64(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a numeric value"}
		}

		if v <= 0 {
			return &RuleError{Code: ErrMin, Message: "must be greater than 0"}
		}

		return nil
	}
}

// NonNegative fails when value is less than 0.
func NonNegative() Rule {
	return func(value any) error {
		v, ok := toFloat64(value)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "expected a numeric value"}
		}

		if v < 0 {
			return &RuleError{Code: ErrMin, Message: "must be 0 or greater"}
		}

		return nil
	}
}

// --- generic rules ---

// OneOf fails when value is not among the provided allowed values.
func OneOf[T comparable](values ...T) Rule {
	return func(value any) error {
		v, ok := value.(T)
		if !ok {
			return &RuleError{Code: ErrInvalid, Message: "invalid value"}
		}

		if slices.Contains(values, v) {
			return nil
		}

		return &RuleError{Code: ErrInvalid, Message: "must be one of the allowed values"}
	}
}

// Optional wraps rules and skips them when value is nil or whitespace-only.
// Non-empty values are passed through all rules in order; first failure stops.
func Optional(rules ...Rule) Rule {
	return func(value any) error {
		if isEmpty(value) {
			return nil
		}

		for _, rule := range rules {
			if err := rule(value); err != nil {
				return err
			}
		}

		return nil
	}
}

// --- helpers ---

// asString returns the string value and true for string and *string types.
// Returns "", false for all other types including untyped nil.
func asString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case *string:
		if v == nil {
			return "", true
		}

		return *v, true
	default:
		return "", false
	}
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case *string:
		return v == nil || strings.TrimSpace(*v) == ""
	default:
		return false
	}
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func containsSpecial(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return true
		}
	}

	return false
}
