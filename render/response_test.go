package render_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/patrickward/hop/v2/alert"
	"github.com/patrickward/hop/v2/render"
	"github.com/patrickward/hop/v2/render/testdata/templates"
)

func TestResponseAddsFlashManager(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	flash := alert.NewFlashManager("flash", &sessionManagerMock{})
	rs := render.NewResponse(createResponseTemplateManager(t)).WithFlash(flash)
	flash.Add(ctx, alert.TypeSuccess, "Flash message added")
	flash.AddAlert(ctx, "This is an alert message")

	td := rs.ToTemplateData(createSimpleRequest(t))
	if len(td.Flash) != 2 {
		t.Errorf("Expected 2 flash messages, got %d", len(td.Flash))
	}

	if td.Flash[0].Type != alert.TypeSuccess || td.Flash[0].Content != "Flash message added" {
		t.Errorf("Expected flash message type %s with content 'Flash message added', got type %s with content '%s'",
			alert.TypeSuccess, td.Flash[0].Type, td.Flash[0].Content)
	}

	if td.Flash[1].Type != alert.TypeAlert || td.Flash[1].Content != "This is an alert message" {
		t.Errorf("Expected flash message type %s with content 'This is an alert message', got type %s with content '%s'",
			alert.TypeAlert, td.Flash[1].Type, td.Flash[1].Content)
	}
}

func TestResponseSetsLayout(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Layout("custom_layout")

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Layout != "custom_layout" {
		t.Errorf("Expected layout 'custom_layout', got '%s'", td.Layout)
	}
}

func TestResponseSetsPath(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Path("custom_path")

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Path != "pages/custom_path" {
		t.Errorf("Expected path 'pages/custom_path', got '%s'", td.Path)
	}
}

func TestResponseSetsTitle(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Title("Custom Title")

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Title != "Custom Title" {
		t.Errorf("Expected title 'Custom Title', got '%s'", td.Title)
	}
}

func TestResponseSetsData(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Data("key1", "value1").Data("key2", 42)

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Data["key1"] != "value1" {
		t.Errorf("Expected data key1 'value1', got '%s'", td.Data["key1"])
	}
	if td.Data["key2"] != 42 {
		t.Errorf("Expected data key2 42, got '%v'", td.Data["key2"])
	}
}

func TestResponseMergesData(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Data("key1", "value1")
	rs.MergeData(map[string]any{"key2": 42})

	td := rs.ToTemplateData(createSimpleRequest(t))
	if td.Data["key1"] != "value1" {
		t.Errorf("Expected data key1 'value1', got '%s'", td.Data["key1"])
	}

	if td.Data["key2"] != 42 {
		t.Errorf("Expected data key2 42, got '%v'", td.Data["key2"])
	}
}

func TestResponseResetsData(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Data("key1", "value1")
	rs.ResetData(map[string]any{"key2": 42})

	td := rs.ToTemplateData(createSimpleRequest(t))
	if td.Data["key1"] != nil {
		t.Errorf("Expected data key1 to be reset, got '%s'", td.Data["key1"])
	}

	if td.Data["key2"] != 42 {
		t.Errorf("Expected data key2 42, got '%v'", td.Data["key2"])
	}
}

func TestResponseAddsHeaders(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.AddHeader("X-Custom-Header", "CustomValue").AddHeader("Content-Type", "application/json")

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Headers["X-Custom-Header"] != "CustomValue" {
		t.Errorf("Expected header 'X-Custom-Header' to be 'CustomValue', got '%s'", td.Headers["X-Custom-Header"])
	}
	if td.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected header 'Content-Type' to be 'application/json', got '%s'", td.Headers["Content-Type"])
	}
}

func TestResponseResetsHeaders(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.AddHeader("X-Custom-Header", "CustomValue")
	rs.ResetHeaders(map[string]string{"Content-Type": "application/json"})

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Headers["X-Custom-Header"] != "" {
		t.Errorf("Expected header 'X-Custom-Header' to be reset, got '%s'", td.Headers["X-Custom-Header"])
	}
	if td.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected header 'Content-Type' to be 'application/json', got '%s'", td.Headers["Content-Type"])
	}
}

func TestResponseSetsFieldErrors(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.FieldError("field1", "error1").FieldError("field2", "error2")

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.FieldErrors["field1"] != "error1" {
		t.Errorf("Expected field error for field1 'error1', got '%s'", td.FieldErrors["field1"])
	}
	if td.FieldErrors["field2"] != "error2" {
		t.Errorf("Expected field error for field2 'error2', got '%s'", td.FieldErrors["field2"])
	}
}

func TestResponseResetsFieldErrors(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.FieldError("field1", "error1")
	rs.ResetFieldErrors(map[string]string{"field2": "error2"})

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.FieldErrors["field1"] != "" {
		t.Errorf("Expected field error for field1 to be reset, got '%s'", td.FieldErrors["field1"])
	}
	if td.FieldErrors["field2"] != "error2" {
		t.Errorf("Expected field error for field2 'error2', got '%s'", td.FieldErrors["field2"])
	}
}

func TestResponseResetsFieldErrorsToNil(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.FieldError("field1", "error1")
	rs.ResetFieldErrors(nil)

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.FieldErrors["field1"] != "" {
		t.Errorf("Expected field error for field1 to be reset, got '%s'", td.FieldErrors["field1"])
	}
	if len(td.FieldErrors) != 0 {
		t.Errorf("Expected no field errors, got %d", len(td.FieldErrors))
	}
}

func TestResponseSetsNonce(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Nonce("random-nonce-value")

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Nonce != "random-nonce-value" {
		t.Errorf("Expected nonce 'random-nonce-value', got '%s'", td.Nonce)
	}
}

// TestResponseSetsMeta tests setting meta data in the response
func TestResponseSetsMeta(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Meta("key1", "value1").Meta("key2", 42)

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Meta["key1"] != "value1" {
		t.Errorf("Expected meta key1 'value1', got '%s'", td.Meta["key1"])
	}
	if td.Meta["key2"] != 42 {
		t.Errorf("Expected meta key2 42, got '%v'", td.Meta["key2"])
	}
}

// TestResponseResetsMeta tests resetting meta data in the response
func TestResponseResetsMeta(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Meta("key1", "value1")
	rs.ResetMeta(map[string]any{"key2": 42})

	td := rs.ToTemplateData(createSimpleRequest(t))

	if td.Meta["key1"] != nil {
		t.Errorf("Expected meta key1 to be reset, got '%s'", td.Meta["key1"])
	}
	if td.Meta["key2"] != 42 {
		t.Errorf("Expected meta key2 42, got '%v'", td.Meta["key2"])
	}
}

func TestResponseSetsMessages(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Message(alert.TypeInfo, "This is an info message")
	rs.Message(alert.TypeError, "This is an error message")

	td := rs.ToTemplateData(createSimpleRequest(t))

	if len(td.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(td.Messages))
	}

	if td.Messages[0].Type != "info" || td.Messages[0].Content != "This is an info message" {
		t.Errorf("Expected first message to be info, got %s: %s", td.Messages[0].Type, td.Messages[0].Content)
	}

	if td.Messages[1].Type != "error" || td.Messages[1].Content != "This is an error message" {
		t.Errorf("Expected second message to be error, got %s: %s", td.Messages[1].Type, td.Messages[1].Content)
	}
}

func TestResponseResetsMessages(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.Message(alert.TypeInfo, "This is an info message")
	rs.ResetMessages(alert.Messages{
		{Type: alert.TypeSuccess, Content: "This is a success message"},
	})

	td := rs.ToTemplateData(createSimpleRequest(t))

	if len(td.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(td.Messages))
	}

	if td.Messages[0].Type != "success" || td.Messages[0].Content != "This is a success message" {
		t.Errorf("Expected message to be success, got %s: %s", td.Messages[0].Type, td.Messages[0].Content)
	}
}

func TestResponseWriteRendersTemplates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		setupResponse   func(*render.Response)
		expectedStatus  int
		expectedContent []string
		expectedHeaders map[string]string
	}{
		{
			name: "basic home page render",
			setupResponse: func(rs *render.Response) {
				rs.Path("home").
					Title("Home Page").
					Data("Content", "Home Page TEST 0001").
					Data("User", "Wile Coyote")
			},
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"<title>Home Page</title>",
				`<header class="main-header">`,
				"Home Page TEST 0001",
				"Wile Coyote",
			},
		},
		{
			name: "render with a different layout and partials",
			setupResponse: func(rs *render.Response) {
				rs.Layout("admin").
					Path("admin/dashboard").
					Title("Admin Dashboard").
					Data("Content", "Admin Dashboard Content")
			},
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"<title>Admin Dashboard</title>",
				`<header class="admin-header">`,
				`<h1 class="admin-title">Admin Dashboard</h1>`,
				`<div class="dashboard-content">Admin Dashboard Content</div>`,
			},
		},
		{
			name: "page with custom headers",
			setupResponse: func(rs *render.Response) {
				rs.Path("home").
					Title("Home").
					Data("Content", "Page With Headers").
					AddHeader("X-Custom-Header", "CustomValue")
			},
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"<title>Home</title>",
				`<header class="main-header">`,
				"Page With Headers",
			},
			expectedHeaders: map[string]string{
				"X-Custom-Header": "CustomValue",
			},
		},
		{
			name: "page with field errors",
			setupResponse: func(rs *render.Response) {
				rs.Path("form").
					Title("Form Page").
					Data("Content", "Form with Errors").
					FieldError("name", "Name is required").
					FieldError("email", "Email is invalid")
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedContent: []string{
				"<title>Form Page</title>",
				"Form with Errors",
				`<div class="form-group error">`,
				`<div class="error-message">Name is required</div>`,
				`<div class="error-message">Email is invalid</div>`,
			},
		},
		{
			name: "page with messages",
			setupResponse: func(rs *render.Response) {
				rs.Path("home").
					Title("Messages Page").
					Data("Content", "Page with Messages").
					Message(alert.TypeInfo, "This is an info message").
					Message(alert.TypeError, "This is an error message")
			},
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"<title>Messages Page</title>",
				`<header class="main-header">`,
				"Page with Messages",
				`<div class="message info">This is an info message</div>`,
				`<div class="message error">This is an error message</div>`,
			},
		},
		{
			name: "page with flash messages",
			setupResponse: func(rs *render.Response) {
				rs.Path("home").
					Title("Flash Messages Page").
					Data("Content", "Page with Flash Messages")
			},
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"<title>Flash Messages Page</title>",
				`<header class="main-header">`,
				"Page with Flash Messages",
				`<div class="flash notice">This is a notice message</div>`,
				`<div class="flash alert">This is an alert message</div>`,
			},
		},
		{
			name: "page with htmx trigger",
			setupResponse: func(rs *render.Response) {
				rs.Path("home").
					Title("HTMX Trigger Page").
					Data("Content", "Page with HTMX Trigger").
					HxTrigger("foo-bar", "bar-baz")
			},
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"<title>HTMX Trigger Page</title>",
				`<header class="main-header">`,
				"Page with HTMX Trigger",
			},
			expectedHeaders: map[string]string{
				"HX-Trigger": `{"foo-bar":"bar-baz"}`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := createSimpleRequest(t)

			flash := alert.NewFlashManager("flash", &sessionManagerMock{})
			rs := render.NewResponse(createResponseTemplateManager(t)).WithFlash(flash)

			if tc.name == "page with flash messages" {
				flash.AddNotice(r.Context(), "This is a notice message")
				flash.AddAlert(r.Context(), "This is an alert message")
			}

			tc.setupResponse(rs)

			err := rs.Write(w, r)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, w.Code)
			}

			if tc.expectedHeaders != nil {
				for key, expectedValue := range tc.expectedHeaders {
					if value := w.Header().Get(key); value != expectedValue {
						t.Errorf("Expected header '%s' to be '%s', got '%s'", key, expectedValue, value)
					}
				}
			}

			body := w.Body.String()
			for _, content := range tc.expectedContent {
				if !strings.Contains(body, content) {
					t.Errorf("Expected response body to contain '%s', got '%s'", content, body)
				}
			}
		})
	}
}

// createResponseTemplateManager creates a new TemplateManager for testing purposes
func createResponseTemplateManager(t *testing.T) *render.TemplateManager {
	t.Helper()

	return render.NewTemplateManager(templates.FS, render.WithExtension(".gtml"))
}

func createSimpleRequest(t *testing.T) *http.Request {
	t.Helper()

	r, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	return r
}

//func createRequest(t *testing.T, method, url string) *http.Request {
//	t.Helper()
//
//	r, err := http.NewRequest(method, url, nil)
//	if err != nil {
//		t.Fatalf("Failed to create request: %v", err)
//	}
//	return r
//}
