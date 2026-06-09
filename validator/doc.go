// Package validator provides a lightweight, composable field validator.
//
// Build a [Validator], call [Validator.Field] for each input field with one or more [Rule]
// functions, then call [Validator.Result] to retrieve a [*ValidationError] (nil when all
// fields pass).
//
// Example:
//
//	v := validator.New()
//	v.Field("email", req.Email, validator.Required(), validator.Email())
//	v.Field("age", req.Age, validator.Gte(18))
//	if err := v.Result(); err != nil {
//	    // err.Fields contains per-field details
//	}
package validator
