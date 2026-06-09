// Package errs provides a structured, transport-agnostic error type for
// domain and application layers.
//
// Errors carry a machine-readable Code, a human-readable Message, optional
// per-field validation failures, and optional internal metadata. They contain
// no HTTP-specific concepts — the HTTP layer maps codes to status codes.
//
// # Usage
//
//	// standard constructors
//	errs.NotFound("user not found")
//	errs.InvalidInput("bad request", errs.NewFieldError("email", errs.CodeRequired, "required"))
//
//	// app-specific codes
//	errs.New("EMAIL_ALREADY_EXISTS", "email is already taken")
//
//	// attach internal context (never exposed to clients)
//	errs.Internal("unexpected error").WithCause(dbErr).WithMeta("user_id", id)
//
//	// errors.As traversal works through fmt.Errorf wrapping
//	wrapped := fmt.Errorf("UserService.GetByID: %w", errs.NotFound("user not found"))
//	var e *errs.Error
//	errors.As(wrapped, &e) // true
package errs
