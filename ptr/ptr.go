package ptr

// Of returns a pointer to the given value.
func Of[T any](v T) *T {
	return &v
}

// ValueOf dereferences p and returns the value.
// If p is nil, fallback is returned instead.
func ValueOf[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}

	return *p
}

// Deref dereferences p and returns the value, or the zero value of T if p is nil.
func Deref[T any](p *T) T {
	var zero T

	if p == nil {
		return zero
	}

	return *p
}

// IsNil returns true if p is nil.
func IsNil[T any](p *T) bool {
	return p == nil
}

// Equal reports whether two pointers are equal.
// Both nil pointers are considered equal.
// Two non-nil pointers are equal when the values they point to are equal.
func Equal[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}
