package validator

import "errors"

// Rule is a validation function. Returns nil when the value is valid.
// On failure, return a *RuleError to carry a code and message.
type Rule func(value any) error

// Validator accumulates field errors for a request payload.
type Validator struct {
	fields []FieldError
}

// New creates a new Validator.
func New() *Validator {
	return &Validator{}
}

// Field runs rules against value in order, recording the first failure for fieldName.
// Subsequent rules are skipped once a failure is found.
// Rules that return a plain error (not *RuleError) are recorded with CodeInvalid.
func (v *Validator) Field(fieldName string, value any, rules ...Rule) {
	for _, rule := range rules {
		if err := rule(value); err != nil {
			if re, ok := errors.AsType[*RuleError](err); ok {
				v.fields = append(v.fields, FieldError{
					Field:   fieldName,
					Code:    re.Code,
					Message: re.Message,
					Params:  re.Params,
				})
			} else {
				v.fields = append(v.fields, FieldError{
					Field:   fieldName,
					Code:    ErrInvalid,
					Message: err.Error(),
				})
			}

			return
		}
	}
}

// Result returns nil if no errors were recorded, or a *ValidationError with all field errors.
func (v *Validator) Result() *ValidationError {
	if len(v.fields) == 0 {
		return nil
	}

	return &ValidationError{Message: "validation failed", Fields: v.fields}
}
