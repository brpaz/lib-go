package nullable

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Value holds an optional value of type T.
type Value[T any] struct {
	Val   T
	Valid bool // true if Val is set
}

// Of returns a valid Value wrapping v.
func Of[T any](v T) Value[T] {
	return Value[T]{Val: v, Valid: true}
}

// Empty returns an invalid (null) Value of type T.
func Empty[T any]() Value[T] {
	return Value[T]{}
}

// FromPtr converts a pointer to a Value. A nil pointer becomes an invalid
// (null) Value; a non-nil pointer becomes valid, wrapping a copy of the
// pointed-to value. Useful at boundaries with pointer-based APIs (generated
// DTOs, ORMs) that represent optional fields as *T.
func FromPtr[T any](p *T) Value[T] {
	if p == nil {
		return Empty[T]()
	}

	return Of(*p)
}

// Get returns (value, true) if valid, or (zero, false) if not.
func (n Value[T]) Get() (T, bool) {
	return n.Val, n.Valid
}

// ValueOr returns the value if valid, otherwise fallback.
func (n Value[T]) ValueOr(fallback T) T {
	if n.Valid {
		return n.Val
	}

	return fallback
}

// ToPtr converts n to a pointer. An invalid Value becomes nil; a valid
// Value becomes a pointer to a copy of Val. Pairs with [FromPtr] for
// round-tripping through pointer-based APIs.
func (n Value[T]) ToPtr() *T {
	if !n.Valid {
		return nil
	}

	v := n.Val

	return &v
}

// IsNull returns true if the value is not valid.
func (n Value[T]) IsNull() bool {
	return !n.Valid
}

// String implements fmt.Stringer, returning a textual representation of Val
// when valid, or "<null>" when not — useful for logging and debugging.
func (n Value[T]) String() string {
	if !n.Valid {
		return "<null>"
	}

	return fmt.Sprint(n.Val)
}

// Set sets the value and marks it valid.
func (n *Value[T]) Set(v T) {
	n.Val = v
	n.Valid = true
}

// Clear marks the Value as invalid/null.
func (n *Value[T]) Clear() {
	var zero T
	n.Val = zero
	n.Valid = false
}

// MarshalJSON marshals to the JSON value if valid, or null if not.
func (n Value[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.Val)
}

// UnmarshalJSON parses JSON null as invalid, otherwise unmarshals into Val and sets Valid=true.
func (n *Value[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		n.Clear()
		return nil
	}

	if err := json.Unmarshal(data, &n.Val); err != nil {
		return err
	}

	n.Valid = true

	return nil
}

// MarshalText implements encoding.TextMarshaler. It returns an empty byte slice
// when invalid, or the textual representation of Val when valid — the form
// expected by JSON map keys, url.Values and form/query encoders. T must be a
// string, a basic numeric/bool kind, or implement [encoding.TextMarshaler];
// any other type returns an error.
func (n Value[T]) MarshalText() ([]byte, error) {
	if !n.Valid {
		return []byte{}, nil
	}

	if marshaler, ok := any(n.Val).(encoding.TextMarshaler); ok {
		return marshaler.MarshalText()
	}

	rv := reflect.ValueOf(n.Val)

	switch rv.Kind() {
	case reflect.String:
		return []byte(rv.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(rv.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(rv.Uint(), 10)), nil
	case reflect.Float32, reflect.Float64:
		return []byte(strconv.FormatFloat(rv.Float(), 'f', -1, 64)), nil
	case reflect.Bool:
		return []byte(strconv.FormatBool(rv.Bool())), nil
	default:
		return nil, fmt.Errorf("nullable: cannot marshal value of type %T to text", n.Val)
	}
}

// UnmarshalText implements encoding.TextUnmarshaler. Empty input marks the
// Value as invalid (null); otherwise it parses the text into T. T must be a
// string, a basic numeric/bool kind, or *T must implement
// [encoding.TextUnmarshaler]; any other type returns an error.
func (n *Value[T]) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		n.Clear()
		return nil
	}

	if unmarshaler, ok := any(&n.Val).(encoding.TextUnmarshaler); ok {
		if err := unmarshaler.UnmarshalText(text); err != nil {
			return err
		}

		n.Valid = true

		return nil
	}

	rv := reflect.ValueOf(&n.Val).Elem()
	s := string(text)

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("nullable: cannot unmarshal %q into %T: %w", s, n.Val, err)
		}

		rv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("nullable: cannot unmarshal %q into %T: %w", s, n.Val, err)
		}

		rv.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("nullable: cannot unmarshal %q into %T: %w", s, n.Val, err)
		}

		rv.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("nullable: cannot unmarshal %q into %T: %w", s, n.Val, err)
		}

		rv.SetBool(b)
	default:
		return fmt.Errorf("nullable: cannot unmarshal text into %T", n.Val)
	}

	n.Valid = true

	return nil
}

// Scan implements sql.Scanner, allowing Value to be used as a scan destination.
// A nil src sets Valid=false; otherwise it delegates to sql.Scanner if T implements it,
// or falls back to reflect-based assignment, and sets Valid=true on success.
func (n *Value[T]) Scan(src any) error {
	if src == nil {
		n.Clear()
		return nil
	}

	if scanner, ok := any(&n.Val).(sql.Scanner); ok {
		if err := scanner.Scan(src); err != nil {
			return err
		}

		n.Valid = true

		return nil
	}

	rv := reflect.ValueOf(&n.Val).Elem()
	sv := reflect.ValueOf(src)

	if sv.Type().AssignableTo(rv.Type()) {
		rv.Set(sv)
	} else if sv.Type().ConvertibleTo(rv.Type()) {
		rv.Set(sv.Convert(rv.Type()))
	} else {
		return fmt.Errorf("nullable: cannot scan value of type %T into %T", src, n.Val)
	}

	n.Valid = true

	return nil
}

// Value implements driver.Valuer, returning nil if not valid, or the underlying value if valid.
func (n Value[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	if valuer, ok := any(n.Val).(driver.Valuer); ok {
		return valuer.Value()
	}

	return n.Val, nil
}

// Equal reports whether a and b represent the same optional value.
// Two invalid Values are equal; a valid and an invalid are not;
// two valid Values are equal when their underlying values are equal.
func Equal[T comparable](a, b Value[T]) bool {
	if a.Valid != b.Valid {
		return false
	}

	if !a.Valid {
		return true
	}

	return a.Val == b.Val
}
