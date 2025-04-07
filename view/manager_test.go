package view_test

import (
	"html/template"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/v2/view"
	"github.com/patrickward/hop/v2/view/testdata/source1"
)

type TestData struct {
	Title      string
	Content    string
	User       string
	IsAdmin    bool
	Navigation []string
}

func (td TestData) toMap() map[string]interface{} {
	return map[string]interface{}{
		"Title":      td.Title,
		"Content":    td.Content,
		"User":       td.User,
		"IsAdmin":    td.IsAdmin,
		"Navigation": td.Navigation,
	}
}

func TestTemplateManager(t *testing.T) {
	tests := []struct {
		name           string
		layout         string
		page           string
		data           TestData
		requestPath    string
		requestMethod  string
		requestHeaders map[string]string
		expectedStatus int
		expectedParts  []string
		expectError    bool
	}{
		{
			name:   "basic page with base layout",
			layout: "base",
			page:   "home",
			data: TestData{
				Title:      "Welcome Home",
				Content:    "Main content here",
				User:       "John Doe",
				IsAdmin:    false,
				Navigation: []string{"Home", "About", "Contact"},
			},
			requestPath:   "/",
			requestMethod: "GET",
			requestHeaders: map[string]string{
				"Accept": "text/html",
			},
			expectedStatus: 200,
			expectedParts: []string{
				"<title>Welcome Home</title>",
				"Main content here",
				"John Doe",
				"<nav>Home</nav>",
			},
			expectError: false,
		},
		{
			name:   "admin page with admin layout",
			layout: "admin",
			page:   "admin/dashboard",
			data: TestData{
				Title:      "Admin Dashboard",
				Content:    "Dashboard Content",
				User:       "Admin User",
				IsAdmin:    true,
				Navigation: []string{"Dashboard", "Users", "Settings"},
			},
			requestPath:   "/admin/dashboard",
			requestMethod: "GET",
			requestHeaders: map[string]string{
				"Authorization": "Bearer test-token",
			},
			expectedStatus: 200,
			expectedParts: []string{
				"<title>Admin Dashboard</title>",
				"Dashboard Content",
				"Admin User",
				"<div class=\"admin-header\">",
			},
			expectError: false,
		},
		{
			name:   "clean layout",
			layout: "clean",
			page:   "home",
			data: TestData{
				Title:   "Clean Layout",
				Content: "Minimal content",
			},
			requestPath:   "/clean",
			requestMethod: "GET",
			requestHeaders: map[string]string{
				"Accept": "text/html",
			},
			expectedStatus: 200,
			expectedParts: []string{
				"<main class=\"clean-layout\">",
				"Minimal content",
			},
			expectError: false,
		},
		{
			name:           "non-existent layout",
			layout:         "missing",
			page:           "home",
			data:           TestData{},
			requestPath:    "/",
			requestMethod:  "GET",
			requestHeaders: map[string]string{},
			expectedStatus: 500,
			expectedParts: []string{
				"layout:missing",
			},
			expectError: true,
		},
		{
			name:           "non-existent page",
			layout:         "base",
			page:           "missing",
			data:           TestData{},
			requestPath:    "/missing",
			requestMethod:  "GET",
			requestHeaders: map[string]string{},
			expectedStatus: 404,
			expectedParts: []string{
				"template not found",
			},
			expectError: false,
		},
	}

	logger := slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil))
	funcMap := template.FuncMap{
		"isAdmin": func(user string) bool {
			return strings.HasPrefix(user, "Admin")
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize template manager
			tm, err := view.NewTemplateManager(
				source1.FS,
				view.TemplateManagerOptions{
					Extension: ".gtml",
					Funcs:     funcMap,
					Logger:    logger,
				})

			require.NoError(t, err, "Failed to create template manager")

			// Test rendering
			req := httptest.NewRequest(tt.requestMethod, tt.requestPath, nil)
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			err = tm.NewResponse(nil).
				Layout(tt.layout).
				Path(tt.page).
				SetData(tt.data.toMap()).
				Title(tt.data.Title).
				Write(w, req)

			result := w.Body.String()

			if tt.expectError {
				assert.Contains(t, result, "error", "Expected error in response")
				return
			}

			assert.Equal(t, tt.expectedStatus, w.Code, "Expected status code %d, got %d", tt.expectedStatus, w.Code)

			for _, expected := range tt.expectedParts {
				assert.Contains(t, result, expected,
					"Expected content not found in rendered template. Got %s. Want %s", result, expected)
			}
		})
	}
}

// TestTemplateManagerInitialization tests the initialization of the template manager
func TestTemplateManagerInitialization(t *testing.T) {
	tests := []struct {
		name        string
		extension   string
		expectError bool
	}{
		{
			name:        "valid extension",
			extension:   ".gtml",
			expectError: false,
		},
		{
			name:        "extension without dot",
			extension:   "gtml",
			expectError: false,
		},
		{
			name:        "empty extension defaults to .html",
			extension:   "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := view.NewTemplateManager(
				source1.FS,
				view.TemplateManagerOptions{
					Extension: tt.extension,
				})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
