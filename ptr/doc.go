// Package ptr provides generic helper functions for working with pointer values.
//
// # Taking a pointer
//
// [Of] returns a pointer to any value inline, avoiding the temporary-variable
// boilerplate required when an API demands a *T:
//
//	p := ptr.Of(42)       // *int pointing to 42
//	s := ptr.Of("hello")  // *string
//
// # Dereferencing safely
//
// [ValueOf] dereferences a pointer and returns a fallback when the pointer is nil:
//
//	var p *int
//	n := ptr.ValueOf(p, 0)  // 0  — p is nil
//	p  = ptr.Of(7)
//	n  = ptr.ValueOf(p, 0)  // 7
//
// [Deref] is the shorthand for the common case where the fallback is just T's
// zero value:
//
//	var p *int
//	n := ptr.Deref(p)  // 0  — p is nil
//	p  = ptr.Of(7)
//	n  = ptr.Deref(p)  // 7
//
// # Nil checks
//
// [IsNil] is a readable nil guard:
//
//	if ptr.IsNil(cfg.Timeout) {
//	    // field was not set
//	}
//
// # Equality
//
// [Equal] compares two pointers nil-safely by comparing the pointed-to values:
//
//	ptr.Equal(ptr.Of("foo"), ptr.Of("foo")) // true
//	ptr.Equal(nil, nil)                     // true  — both nil
//	ptr.Equal(ptr.Of("foo"), nil)           // false — one nil
package ptr
