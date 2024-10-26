package check_test

import (
	"encoding/json"
	"testing"

	"github.com/patrickward/hop/check"
)

// TODO: Add more tests for validator; go through each method and ensure coverage an edge cases

// TestNew ensures a new validator is properly initialized
func TestNew(t *testing.T) {
	v := check.New()

	if v.HasErrors() {
		t.Error("New validator should not have errors")
	}
}

// TestCheck tests standalone error validation
func TestCheck(t *testing.T) {
	tests := []struct {
		name    string
		valid   bool
		message string
		want    bool
	}{
		{
			name:    "valid check",
			valid:   true,
			message: "should not see this",
			want:    true,
		},
		{
			name:    "invalid check",
			valid:   false,
			message: "error occurred",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := check.New()
			got := v.Check(tt.valid, tt.message)

			if got != tt.want {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}

			if !tt.valid && !v.HasErrors() {
				t.Error("HasErrors() should be true for invalid check")
			}
		})
	}
}

// TestCheckField tests field-specific validation
func TestCheckField(t *testing.T) {
	tests := []struct {
		name    string
		valid   bool
		field   string
		message string
		want    bool
	}{
		{
			name:    "valid field",
			valid:   true,
			field:   "username",
			message: "should not see this",
			want:    true,
		},
		{
			name:    "invalid field",
			valid:   false,
			field:   "email",
			message: "invalid email",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := check.New()
			got := v.CheckField(tt.valid, tt.field, tt.message)

			if got != tt.want {
				t.Errorf("CheckField() = %v, want %v", got, tt.want)
			}

			if !tt.valid {
				if !v.HasField(tt.field) {
					t.Errorf("HasField(%q) should be true", tt.field)
				}

				if msg := v.Field(tt.field); msg != tt.message {
					t.Errorf("Field(%q) = %q, want %q", tt.field, msg, tt.message)
				}
			}
		})
	}
}

// TestMultipleErrors tests handling multiple errors
func TestMultipleErrors(t *testing.T) {
	v := check.New()

	// Add multiple field errors
	v.CheckField(false, "username", "username required")
	v.CheckField(false, "username", "username too short")
	v.CheckField(false, "email", "invalid email")

	// Add standalone error
	v.Check(false, "general error")

	// Test field error count
	if !v.HasField("username") {
		t.Error("HasField(username) should be true")
	}

	// Test first error message for field
	if msg := v.Field("username"); msg != "username required" {
		t.Errorf("Field(username) = %q, want %q", msg, "username required")
	}

	// Test all fields have errors
	fields := v.Fields()
	if len(fields) != 2 { // username and email
		t.Errorf("Fields() returned %d fields, want 2", len(fields))
	}
}

// TestJSON tests JSON marshaling of validator
func TestJSON(t *testing.T) {
	v := check.New()

	v.CheckField(false, "username", "invalid username")
	v.Check(false, "general error")

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal validator: %v", err)
	}

	// Define expected structure
	type jsonResponse struct {
		Fields map[string]string `json:"fields,omitempty"`
		Errors []string          `json:"errors,omitempty"`
	}

	var response jsonResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if msg, ok := response.Fields["username"]; !ok || msg != "invalid username" {
		t.Errorf("JSON fields = %v, want username:invalid username", response.Fields)
	}

	// Verify standalone errors
	if len(response.Errors) != 1 || response.Errors[0] != "general error" {
		t.Errorf("JSON errors = %v, want [general error]", response.Errors)
	}
}

// TestClear tests clearing all errors
func TestClear(t *testing.T) {
	v := check.New()

	v.CheckField(false, "field1", "error1")
	v.Check(false, "error2")

	if !v.HasErrors() {
		t.Error("Validator should have errors before Clear()")
	}

	v.Clear()

	if v.HasErrors() {
		t.Error("Validator should not have errors after Clear()")
	}

	if v.HasField("field1") {
		t.Error("Cleared validator should not have field errors")
	}
}

// TestError tests the Error() string output
func TestError(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*check.Validator)
		expected string
	}{
		{
			name: "single field error",
			setup: func(v *check.Validator) {
				v.CheckField(false, "username", "invalid username")
			},
			expected: "username: invalid username",
		},
		{
			name: "multiple errors",
			setup: func(v *check.Validator) {
				v.CheckField(false, "username", "invalid username")
				v.Check(false, "general error")
			},
			expected: "general error; username: invalid username",
		},
		{
			name:     "no errors",
			setup:    func(v *check.Validator) {},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := check.New()
			tt.setup(v)

			if got := v.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestValidatorInTemplateContext tests the validator's template-friendly methods
func TestValidatorInTemplateContext(t *testing.T) {
	v := check.New()

	// Add some errors
	v.CheckField(false, "username", "invalid username")
	v.CheckField(false, "email", "invalid email")

	// Test HasField
	fields := []struct {
		field    string
		expected bool
	}{
		{"username", true},
		{"email", true},
		{"nonexistent", false},
	}

	for _, f := range fields {
		if got := v.HasField(f.field); got != f.expected {
			t.Errorf("HasField(%q) = %v, want %v", f.field, got, f.expected)
		}
	}

	// Test Fields map
	fieldMap := v.Fields()
	if len(fieldMap) != 2 {
		t.Errorf("Fields() returned %d fields, want 2", len(fieldMap))
	}

	expectedFields := map[string]string{
		"username": "invalid username",
		"email":    "invalid email",
	}

	for field, expected := range expectedFields {
		if got := fieldMap[field]; got != expected {
			t.Errorf("Fields()[%q] = %q, want %q", field, got, expected)
		}
	}
}
