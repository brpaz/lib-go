package ptrutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/ptrutil"
)

func TestPtr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    any
		expected any
	}{
		{
			name:     "Int",
			input:    10,
			expected: 10,
		},
		{
			name:     "String",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "Boolean",
			input:    true,
			expected: true,
		},
		{
			name:     "Float",
			input:    3.14,
			expected: 3.14,
		},
		{
			name:     "NilInput",
			input:    nil,
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ptr := ptrutil.Ptr(tc.input)
			assert.NotNil(t, ptr)
			assert.Equal(t, tc.expected, *ptr)
		})
	}
}

func TestDeref(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		ptr      *int
		fallback int
		expected int
	}{
		{
			name:     "NilPointer",
			ptr:      nil,
			fallback: 10,
			expected: 10,
		},
		{
			name:     "ValidPointer",
			ptr:      ptrutil.Ptr(20),
			fallback: 0,
			expected: 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ptrutil.Deref(tc.ptr, tc.fallback)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		ptr      *int
		expected bool
	}{
		{
			name:     "NilPointer",
			ptr:      (*int)(nil),
			expected: true,
		},
		{
			name:     "ValidPointer",
			ptr:      ptrutil.Ptr(10),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ptrutil.IsNil(tc.ptr)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		ptr1     *int
		ptr2     *int
		expected bool
	}{
		{
			name:     "BothNil",
			ptr1:     (*int)(nil),
			ptr2:     (*int)(nil),
			expected: true,
		},
		{
			name:     "OneNil",
			ptr1:     ptrutil.Ptr(10),
			ptr2:     (*int)(nil),
			expected: false,
		},
		{
			name:     "BothEqual",
			ptr1:     ptrutil.Ptr(20),
			ptr2:     ptrutil.Ptr(20),
			expected: true,
		},
		{
			name:     "DifferentValues",
			ptr1:     ptrutil.Ptr(10),
			ptr2:     ptrutil.Ptr(30),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ptrutil.Equal(tc.ptr1, tc.ptr2)
			assert.Equal(t, tc.expected, result)
		})
	}
}
