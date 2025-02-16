package view_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/view"
)

// Test data structures
type TestUser struct {
	ID    string
	Name  string
	Email string
}

func TestPageData(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *view.PageData
		validate func(*testing.T, *view.PageData)
	}{
		{
			name: "basic data operations",
			setup: func() *view.PageData {
				return view.NewPageData().
					SetTitle("Test Page").
					Set("user", TestUser{ID: "1", Name: "John"})
			},
			validate: func(t *testing.T, pd *view.PageData) {
				assert.Equal(t, "Test Page", pd.Title())

				user, ok := pd.Get("user").(TestUser)
				assert.True(t, ok)
				assert.Equal(t, "1", user.ID)
				assert.Equal(t, "John", user.Name)
			},
		},
		{
			name: "merge data operations",
			setup: func() *view.PageData {
				pd := view.NewPageData().SetData(map[string]any{
					"existing": "value",
				})
				pd.Merge(map[string]any{
					"new":      "data",
					"existing": "updated",
				})
				return pd
			},
			validate: func(t *testing.T, pd *view.PageData) {
				assert.Equal(t, "updated", pd.Get("existing"))
				assert.Equal(t, "data", pd.Get("new"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pd := tt.setup()
			tt.validate(t, pd)
		})
	}
}

func TestPageDataMessages(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *view.PageData
		validate func(*testing.T, *view.PageData)
	}{
		{
			name: "validation errors with O(1) lookup",
			setup: func() *view.PageData {
				return view.NewPageData().
					SetError("Please fix the following errors:").
					AddFieldErrors(map[string]string{
						"email": "Invalid email format",
						"name":  "Name is required",
					})
			},
			validate: func(t *testing.T, pd *view.PageData) {
				// Check error message
				messages := pd.MessagesOfType(view.MessageError)
				assert.Len(t, messages, 1)
				assert.Equal(t, "Please fix the following errors:", messages[0].Content)

				// Check field errors - direct lookup
				assert.True(t, pd.HasFieldErrors())
				assert.True(t, pd.HasErrorFor("email"))
				assert.Equal(t, "Invalid email format", pd.ErrorFor("email"))
				assert.Equal(t, "Name is required", pd.ErrorFor("name"))
			},
		},
		{
			name: "adding individual field errors",
			setup: func() *view.PageData {
				return view.NewPageData().
					AddFieldError("username", "Must be at least 3 characters").
					AddFieldError("password", "Too weak")
			},
			validate: func(t *testing.T, pd *view.PageData) {
				assert.True(t, pd.HasFieldErrors())
				assert.Equal(t, "Must be at least 3 characters", pd.ErrorFor("username"))
				assert.Equal(t, "Too weak", pd.ErrorFor("password"))

				// Non-existent field should return empty string
				assert.Equal(t, "", pd.ErrorFor("nonexistent"))
				assert.False(t, pd.HasErrorFor("nonexistent"))
			},
		},
		{
			name: "messages without field errors",
			setup: func() *view.PageData {
				return view.NewPageData().
					SetSuccess("Operation completed").
					SetInfo("Please check your email")
			},
			validate: func(t *testing.T, pd *view.PageData) {
				assert.True(t, pd.HasMessages())
				assert.False(t, pd.HasFieldErrors())

				messages := pd.Messages()
				assert.Len(t, messages, 2)

				successMsgs := pd.MessagesOfType(view.MessageSuccess)
				assert.Len(t, successMsgs, 1)
				assert.Equal(t, "Operation completed", successMsgs[0].Content)
			},
		},
		{
			name: "clearing specific types",
			setup: func() *view.PageData {
				return view.NewPageData().
					SetError("An error occurred").
					AddFieldError("field1", "Invalid").
					ClearFieldErrors().
					ClearMessages()
			},
			validate: func(t *testing.T, pd *view.PageData) {
				assert.False(t, pd.HasMessages())
				assert.False(t, pd.HasFieldErrors())
				assert.Len(t, pd.Messages(), 0)
				assert.Len(t, pd.FieldErrors(), 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pd := tt.setup()
			tt.validate(t, pd)
		})
	}
}

// Example template usage
func ExamplePageData_templateUsage() {
	const tmpl = `
        {{/* Display messages */}}
        {{range .Messages}}
            <div class="alert alert-{{.Type}}">{{.Data.Content}}</div>
        {{end}}

        {{/* Display form with field errors */}}
        <form>
            <div class="form-group">
                <label>Email</label>
                <input name="email" type="email">
                {{if .HasErrorFor "email"}}
                    <span class="error">{{.ErrorFor "email"}}</span>
                {{end}}
            </div>

            {{/* Display all field errors in one place if needed */}}
            {{if .HasFieldErrors}}
                <div class="error-summary">
                    <ul>
                        {{range $field, $error := .FieldErrors}}
                            <li>{{$field}}: {{$error}}</li>
                        {{end}}
                    </ul>
                </div>
            {{end}}
        </form>
    `
	// Output:
}
