package ptrutil

// Ptr creates a pointer to the given value.
func Ptr[T any](value T) *T {
	return &value
}

// Deref safely dereferences a pointer and returns the value.
// If the pointer is nil, it returns the specified default value.
func Deref[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// IsNil checks whether a given pointer is nil.
func IsNil[T any](ptr *T) bool {
	return ptr == nil
}

// Equal checks if two pointers point to the same value or if both are nil.
func Equal[T comparable](ptr1, ptr2 *T) bool {
	if ptr1 == nil && ptr2 == nil {
		return true
	}
	if ptr1 == nil || ptr2 == nil {
		return false
	}
	return *ptr1 == *ptr2
}
