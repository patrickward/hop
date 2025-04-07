package view

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/patrickward/hop/flash"
	"github.com/patrickward/hop/view/htmx"
	"github.com/patrickward/hop/view/htmx/trigger"
)

// Response represents a view response to an HTTP request using html/template
type Response struct {
	status     int
	headers    map[string]string
	layout     string
	path       string
	pageData   *PageData
	tm         *TemplateManager
	flash      *flash.Manager
	triggers   *trigger.Triggers
	extensions []ResponseExtension
}

// NewResponse creates a new Response struct with the provided template manager.
func NewResponse(tm *TemplateManager) *Response {
	return &Response{
		tm:         tm,
		layout:     DefaultBaseLayout,
		status:     http.StatusOK,
		pageData:   NewPageData(),
		headers:    make(map[string]string),
		triggers:   trigger.NewTriggers(),
		extensions: make([]ResponseExtension, 0),
	}
}

// Layout sets the layout template to be used
func (rs *Response) Layout(layout string) *Response {
	rs.layout = layout
	return rs
}

// Path sets the view template path to be used
func (rs *Response) Path(path string) *Response {
	// Handle plugin paths (e.g., "plugin:path")
	if idx := strings.Index(path, ":"); idx >= 0 {
		plugin, rest := path[:idx], path[idx+1:]
		path = plugin + ":" + ViewsDir + "/" + rest
	} else {
		path = ViewsDir + "/" + path
	}
	rs.path = path
	return rs
}

// Title sets the title of the page
func (rs *Response) Title(title string) *Response {
	rs.pageData.SetTitle(title)
	return rs
}

// Data sets a key/value pair to be passed to the template
func (rs *Response) Data(key string, value any) *Response {
	rs.pageData.Set(key, value)
	return rs
}

// SetData resets the data map with the provided data
func (rs *Response) SetData(data map[string]any) *Response {
	rs.pageData.SetData(data)
	return rs
}

// AddHeader sets a header for the response
func (rs *Response) AddHeader(key, value string) *Response {
	rs.headers[key] = value
	return rs
}

// SetHeaders resets all the headers for the response to the provided map
func (rs *Response) SetHeaders(headers map[string]string) *Response {
	rs.headers = headers
	return rs
}

//// WithRequest sets common request related data in the response
//func (rs *Response) WithRequest(r *http.Request) *Response {
//	rs.pageData.Set("CSRFToken", nosurf.Token(r))
//	rs.pageData.Set("RequestPath", r.URL.Path)
//	rs.pageData.Set("IsHome", r.URL.Path == "/")
//	rs.pageData.Set("IsHTMXRequest", htmx.IsHtmxRequest(r))
//	rs.pageData.Set("IsBoostedRequest", htmx.IsBoostedRequest(r))
//	return rs
//}

// Error sets an error message to be displayed on the page
func (rs *Response) Error(msg string) *Response {
	if strings.TrimSpace(msg) != "" {
		rs.status = http.StatusUnprocessableEntity
		rs.pageData.SetError(msg)
	}
	return rs
}

// Success sets a success message to be displayed on the page
func (rs *Response) Success(msg string) *Response {
	if strings.TrimSpace(msg) != "" {
		rs.pageData.SetSuccess(msg)
	}
	return rs
}

// Warning sets a warning message to be displayed on the page
func (rs *Response) Warning(msg string) *Response {
	if strings.TrimSpace(msg) != "" {
		rs.pageData.SetWarning(msg)
	}
	return rs
}

// Info sets an info message to be displayed on the page
func (rs *Response) Info(msg string) *Response {
	if strings.TrimSpace(msg) != "" {
		rs.pageData.SetInfo(msg)
	}
	return rs
}

// FieldErrors sets an error message and field errors to be displayed on the page
func (rs *Response) FieldErrors(msg string, fields map[string]string) *Response {
	rs.status = http.StatusUnprocessableEntity
	rs.pageData.SetError(msg)
	rs.pageData.AddFieldErrors(fields)
	return rs
}

// SetFlashManager sets the flash manager for the response
func (rs *Response) SetFlashManager(f *flash.Manager) *Response {
	rs.flash = f
	return rs
}

// AddExtension adds a ResponseExtension to the response
func (rs *Response) AddExtension(ext ResponseExtension) *Response {
	rs.extensions = append(rs.extensions, ext)
	return rs
}

// RenderNotFound renders a 404 page not found error
func (rs *Response) RenderNotFound(w http.ResponseWriter, r *http.Request) error {
	return rs.handleError(w, r, http.StatusNotFound, fmt.Errorf("page not found"))
}

// RenderForbidden renders a 403 forbidden error
func (rs *Response) RenderForbidden(w http.ResponseWriter, r *http.Request) error {
	return rs.handleError(w, r, http.StatusForbidden, fmt.Errorf("forbidden"))
}

// RenderUnauthorized renders a 401 unauthorized error
func (rs *Response) RenderUnauthorized(w http.ResponseWriter, r *http.Request) error {
	return rs.handleError(w, r, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
}

// RenderSystemError renders a 500 internal server error
func (rs *Response) RenderSystemError(w http.ResponseWriter, r *http.Request, err error) error {
	return rs.handleError(w, r, http.StatusInternalServerError, err)
}

// getHeaders combines regular headers and trigger headers
func (rs *Response) getHeaders() map[string]string {

	if rs.headers == nil {
		rs.headers = make(map[string]string)
	}

	// Add trigger headers if present
	if rs.triggers.HasTriggers() {
		if val, err := rs.triggers.TriggerHeader(); err == nil {
			rs.headers[htmx.HXTrigger] = val
		}
	}

	if rs.triggers.HasAfterSwapTriggers() {
		if val, err := rs.triggers.TriggerAfterSwapHeader(); err == nil {
			rs.headers[htmx.HXTriggerAfterSwap] = val
		}
	}

	if rs.triggers.HasAfterSettleTriggers() {
		if val, err := rs.triggers.TriggerAfterSettleHeader(); err == nil {
			rs.headers[htmx.HXTriggerAfterSettle] = val
		}
	}

	return rs.headers
}

// Write writes the response to the http.ResponseWriter. It implements the ResponseWriter interface.
func (rs *Response) Write(w http.ResponseWriter, r *http.Request) error {
	// Apply extensions first
	for _, ext := range rs.extensions {
		if err := ext.Apply(w, r); err != nil {
			return rs.handleError(w, r, http.StatusInternalServerError, fmt.Errorf("extension error: %w", err))
		}
	}

	// Add all headers to the response, including triggers from htmx
	for k, v := range rs.getHeaders() {
		w.Header().Set(k, v)
	}

	// Add flash messages if available
	if rs.flash != nil {
		if messages := rs.flash.Get(r.Context()); len(messages) > 0 {
			rs.pageData.Set("Flash", messages)
		}
	}

	// Get template
	tmpl, err := rs.tm.getTemplate(rs.path)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrTempNotFound) {
			status = http.StatusNotFound
		}
		return rs.handleError(w, r, status, err)
	}

	// Set status code before writing body
	w.WriteHeader(rs.status)

	// Execute template
	if err := tmpl.ExecuteTemplate(w, fmt.Sprintf("layout:%s", rs.layout), rs.pageData); err != nil {
		return rs.handleError(w, r, http.StatusInternalServerError, fmt.Errorf("template execution error: %w", err))
	}

	return nil
}

// handleError attempts to render an error page, falling back to http.Error if that fails
func (rs *Response) handleError(w http.ResponseWriter, _ *http.Request, status int, originalErr error) error {
	// Log the error
	if rs.tm.logger != nil {
		rs.tm.logger.Error("Template SetError",
			slog.String("path", rs.path),
			slog.String("error", originalErr.Error()))
	}

	// Try to render error template
	errorPath := fmt.Sprintf("%s/%s/%d", ViewsDir, SystemDir, status)
	if errorTmpl, err := rs.tm.getTemplate(errorPath); err == nil {
		rs.pageData.SetTitle(fmt.Sprintf("%d - %s", status, http.StatusText(status))).
			Set("SetError", originalErr.Error()).
			Set("Status", status).
			Set("StatusText", http.StatusText(status))

		w.WriteHeader(status)
		if err := errorTmpl.ExecuteTemplate(w, fmt.Sprintf("layout:%s", rs.tm.systemLayout), rs.pageData); err == nil {
			return nil
		}
	}

	// Fallback to basic error if template fails
	http.Error(w, fmt.Sprintf("%d %s: %v", status, http.StatusText(status), originalErr), status)
	return originalErr
}
