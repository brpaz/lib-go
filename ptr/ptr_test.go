package ptr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/ptr"
)

func TestOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int
	}{
		{name: "zero value", input: 0},
		{name: "positive integer", input: 42},
		{name: "negative integer", input: -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ptr.Of(tc.input)

			require.NotNil(t, got)
			assert.Equal(t, tc.input, *got)
		})
	}
}

func TestValueOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *int
		fallback int
		want     int
	}{
		{
			name:     "non-nil ptr returns dereferenced value",
			input:    ptr.Of(10),
			fallback: 99,
			want:     10,
		},
		{
			name:     "nil ptr returns fallback",
			input:    nil,
			fallback: 99,
			want:     99,
		},
		{
			name:     "zero value ptr returns zero not fallback",
			input:    ptr.Of(0),
			fallback: 99,
			want:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ptr.ValueOf(tc.input, tc.fallback)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDeref(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *int
		want  int
	}{
		{
			name:  "non-nil ptr returns dereferenced value",
			input: ptr.Of(10),
			want:  10,
		},
		{
			name:  "nil ptr returns zero value",
			input: nil,
			want:  0,
		},
		{
			name:  "zero value ptr returns zero",
			input: ptr.Of(0),
			want:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ptr.Deref(tc.input)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input *string
		want  bool
	}{
		{
			name:  "nil ptr returns true",
			input: nil,
			want:  true,
		},
		{
			name:  "non-nil ptr returns false",
			input: ptr.Of("hello"),
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ptr.IsNil(tc.input)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    *int
		b    *int
		want bool
	}{
		{
			name: "both nil are equal",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "same value pointers are equal",
			a:    ptr.Of(5),
			b:    ptr.Of(5),
			want: true,
		},
		{
			name: "different value pointers are not equal",
			a:    ptr.Of(5),
			b:    ptr.Of(6),
			want: false,
		},
		{
			name: "nil and non-nil are not equal",
			a:    nil,
			b:    ptr.Of(5),
			want: false,
		},
		{
			name: "non-nil and nil are not equal",
			a:    ptr.Of(5),
			b:    nil,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ptr.Equal(tc.a, tc.b)

			assert.Equal(t, tc.want, got)
		})
	}
}
