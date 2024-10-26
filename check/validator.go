// Package check provides support for validating data. It can be used to collect errors and field errors.
// Most of this comes from generated code via autostrada.dev and github.com/alexedwards
package check

import "strings"

// Validator is a utility for validating data. It can be used to collect errors and field errors.
type Validator struct {
	Errors      []string
	FieldErrors map[string]string
}

// Error returns the error message for the validator, and implements the error interface.
// It will return a concatenation of all errors and field errors.
func (v *Validator) Error() string {
	var errStr strings.Builder

	for _, err := range v.Errors {
		errStr.WriteString(err)
		errStr.WriteString("; ")
	}

	for key, err := range v.FieldErrors {
		errStr.WriteString(key)
		errStr.WriteString(": ")
		errStr.WriteString(err)
		errStr.WriteString("; ")
	}

	return errStr.String()
}

// HasErrors returns true if there are any errors or field errors.
func (v *Validator) HasErrors() bool {
	return len(v.Errors) != 0 || len(v.FieldErrors) != 0
}

// AddError adds an error message to the validator.
func (v *Validator) AddError(message string) {
	if v.Errors == nil {
		v.Errors = []string{}
	}

	v.Errors = append(v.Errors, message)
}

// AddFieldError adds an error message for a specific field to the validator.
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = map[string]string{}
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// Check adds an error message to the validator if the condition is false.
func (v *Validator) Check(ok bool, message string) {
	if !ok {
		v.AddError(message)
	}
}

// CheckField adds an error message for a specific field to the validator if the condition is false.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}
