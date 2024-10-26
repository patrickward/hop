package check

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type Validator struct {
	fieldErrors map[string][]ValidationError
	errors      []ValidationError
}

// New creates a new validator instance
func New() *Validator {
	return &Validator{
		fieldErrors: make(map[string][]ValidationError),
		errors:      make([]ValidationError, 0),
	}
}

// ensureInitialized initializes internal maps and slices if they're nil
func (v *Validator) ensureInitialized() {
	if v.fieldErrors == nil {
		v.fieldErrors = make(map[string][]ValidationError)
	}
	if v.errors == nil {
		v.errors = make([]ValidationError, 0)
	}
}

// Check performs a validation and adds an error if it fails
func (v *Validator) Check(valid bool, message string) bool {
	if !valid {
		v.AddError(message)
		return false
	}
	return true
}

// CheckField performs a field validation and adds an error if it fails
func (v *Validator) CheckField(valid bool, field, message string) bool {
	if !valid {
		v.AddFieldError(field, message)
		return false
	}
	return true
}

// Error implements the error interface
func (v *Validator) Error() string {
	v.ensureInitialized()

	var sb strings.Builder

	for _, err := range v.errors {
		sb.WriteString(err.Message)
		sb.WriteString("; ")
	}

	for field, errs := range v.fieldErrors {
		if len(errs) > 0 {
			sb.WriteString(fmt.Sprintf("%s: %s; ", field, errs[0].Message))
		}
	}

	return strings.TrimSuffix(sb.String(), "; ")
}

// AddError adds a standalone error
func (v *Validator) AddError(message string) {
	v.ensureInitialized()
	v.errors = append(v.errors, ValidationError{Message: message})
}

// AddFieldError adds an error for a specific field
func (v *Validator) AddFieldError(field, message string) {
	v.ensureInitialized()
	v.fieldErrors[field] = append(v.fieldErrors[field], ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are any validation errors
func (v *Validator) HasErrors() bool {
	v.ensureInitialized()
	return len(v.fieldErrors) > 0 || len(v.errors) > 0
}

// Field returns the first error message for a field if it exists
func (v *Validator) Field(field string) string {
	v.ensureInitialized()
	if errs, exists := v.fieldErrors[field]; exists && len(errs) > 0 {
		return errs[0].Message
	}
	return ""
}

// HasField returns true if the field has any errors
func (v *Validator) HasField(field string) bool {
	v.ensureInitialized()
	errs, exists := v.fieldErrors[field]
	return exists && len(errs) > 0
}

// Fields returns a map of field names to their first error message
func (v *Validator) Fields() map[string]string {
	v.ensureInitialized()
	fields := make(map[string]string)
	for field, errs := range v.fieldErrors {
		if len(errs) > 0 {
			fields[field] = errs[0].Message
		}
	}
	return fields
}

// DetailedErrors returns all validation errors with full details
func (v *Validator) DetailedErrors() map[string][]ValidationError {
	v.ensureInitialized()
	return v.fieldErrors
}

// MarshalJSON implements json.Marshaler
func (v *Validator) MarshalJSON() ([]byte, error) {
	v.ensureInitialized()
	return json.Marshal(struct {
		Fields map[string]string `json:"fields,omitempty"`
		Errors []string          `json:"errors,omitempty"`
	}{
		Fields: v.Fields(),
		Errors: v.messages(),
	})
}

// messages returns all standalone error messages
func (v *Validator) messages() []string {
	v.ensureInitialized()
	msgs := make([]string, len(v.errors))
	for i, err := range v.errors {
		msgs[i] = err.Message
	}
	return msgs
}

// Clear removes all errors from the validator
func (v *Validator) Clear() {
	v.fieldErrors = make(map[string][]ValidationError)
	v.errors = make([]ValidationError, 0)
}
