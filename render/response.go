package render

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/justinas/nosurf"

	"github.com/patrickward/hop/v2/alert"
	"github.com/patrickward/hop/v2/render/htmx"
	"github.com/patrickward/hop/v2/render/htmx/trigger"
)

type Response struct {
	status int
	tm     *TemplateManager
	flash  *alert.FlashManager

	// Template data fields
	layout        string
	path          string
	environment   string
	title         string
	data          map[string]any
	fieldErrors   map[string]string
	headers       map[string]string
	nonce         string
	meta          map[string]any
	messages      alert.Messages
	flashMessages alert.Messages
	triggers      *trigger.Triggers
}

// NewResponse creates a new Response struct with the provided template manager.
func NewResponse(tm *TemplateManager) *Response {
	return &Response{
		status:      http.StatusOK,
		tm:          tm,
		layout:      tm.BaseLayout(),
		data:        make(map[string]any),
		fieldErrors: make(map[string]string),
		headers:     make(map[string]string),
		nonce:       "", // Nonce can be set later if needed
		messages:    alert.Messages{},
		triggers:    trigger.NewTriggers(),
	}
}

// WithFlash sets the flash manager for the response
func (rs *Response) WithFlash(flash *alert.FlashManager) *Response {
	rs.flash = flash
	return rs
}

// Flash returns the flash messages from the flash manager, if set
func (rs *Response) Flash(r *http.Request) alert.Messages {
	if rs.flashMessages == nil && r != nil && rs.flash != nil {
		rs.flashMessages = rs.flash.Pop(r.Context())
	}

	return rs.flashMessages
}

// Layout sets the layout template to be used
func (rs *Response) Layout(layout string) *Response {
	rs.layout = layout
	return rs
}

// Path sets the page template path to be used
func (rs *Response) Path(pathStr string) *Response {
	// Handle plugin paths (e.g., "plugin:path")
	if idx := strings.Index(pathStr, ":"); idx >= 0 {
		plugin, rest := pathStr[:idx], pathStr[idx+1:]
		rs.path = plugin + ":" + path.Join(rs.tm.PagesDir(), rest)
	} else {
		rs.path = path.Join(rs.tm.PagesDir(), pathStr)
	}

	return rs
}

// Environment sets the environment for the response, typically handled by the application
func (rs *Response) Environment(env string) *Response {
	rs.environment = env
	return rs
}

// Title sets the title of the page
func (rs *Response) Title(title string) *Response {
	rs.title = title
	return rs
}

// Data adds a key-value pair to the data map for the template
func (rs *Response) Data(key string, value any) *Response {
	rs.data[key] = value
	return rs
}

// ResetData sets the data for the template
func (rs *Response) ResetData(data map[string]any) *Response {
	rs.data = data
	return rs
}

// MergeData adds a map of data to the existing data map
func (rs *Response) MergeData(data map[string]any) *Response {
	for key, value := range data {
		rs.data[key] = value
	}
	return rs
}

// FieldError adds a field error for the response
func (rs *Response) FieldError(field, message string) *Response {
	rs.status = http.StatusUnprocessableEntity
	rs.fieldErrors[field] = message
	return rs
}

// MergeFieldErrors merges a map of field errors into the existing field errors
func (rs *Response) MergeFieldErrors(fieldErrors map[string]string) *Response {
	for field, message := range fieldErrors {
		rs.status = http.StatusUnprocessableEntity
		rs.fieldErrors[field] = message
	}
	return rs
}

// ResetFieldErrors resets the field errors for the response to the provided map
func (rs *Response) ResetFieldErrors(fieldErrors map[string]string) *Response {
	rs.status = http.StatusUnprocessableEntity
	rs.fieldErrors = fieldErrors
	return rs
}

// AddHeader sets a header for the response
func (rs *Response) AddHeader(key, value string) *Response {
	rs.headers[key] = value
	return rs
}

// ResetHeaders resets all the headers for the response to the provided map
func (rs *Response) ResetHeaders(headers map[string]string) *Response {
	rs.headers = headers
	return rs
}

// Nonce sets the nonce for the response, typically used for Content Security Policy (CSP)
func (rs *Response) Nonce(nonce string) *Response {
	rs.nonce = nonce
	return rs
}

// Meta adds a key-value pair to the meta data for the template
func (rs *Response) Meta(key string, value any) *Response {
	if rs.meta == nil {
		rs.meta = make(map[string]any)
	}
	rs.meta[key] = value
	return rs
}

// ResetMeta resets the meta-data for the response to the provided map
func (rs *Response) ResetMeta(meta map[string]any) *Response {
	rs.meta = meta
	return rs
}

// Message adds a message to the response
func (rs *Response) Message(msgType alert.Type, msg string) *Response {
	if strings.TrimSpace(msg) != "" {
		rs.messages = append(rs.messages, alert.Message{
			Type:    msgType,
			Content: msg,
		})
	}
	return rs
}

// ResetMessages resets the messages for the response to the provided slice
func (rs *Response) ResetMessages(messages alert.Messages) *Response {
	rs.messages = messages
	return rs
}

// CompileHeaders compiles the headers for the response, including any HTMX triggers
func (rs *Response) CompileHeaders() map[string]string {

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

// ToTemplateData converts the Response to a TemplateData object
func (rs *Response) ToTemplateData(r *http.Request) *TemplateData {
	// Set default value for environment
	if rs.environment == "" {
		rs.environment = "development"
	}

	td := &TemplateData{
		Environment: rs.environment,
		Status:      rs.status,
		Headers:     rs.CompileHeaders(),
		Layout:      rs.layout,
		Path:        rs.path,
		Title:       rs.title,
		Data:        make(map[string]any),
		FieldErrors: make(map[string]string),
		Nonce:       rs.nonce,
		Meta:        make(map[string]any),
		Flash:       rs.Flash(r),
		Messages:    rs.messages,
		request:     r,
	}

	// Set the token for CSRF protection if we have a request
	if r != nil {
		td.CSRFToken = nosurf.Token(r)
	}

	// Copy data from the response to the TemplateData
	// to avoid modifying the original data
	for key, value := range rs.data {
		td.Data[key] = value
	}

	// Copy field errors from the response to the TemplateData
	for key, value := range rs.fieldErrors {
		td.FieldErrors[key] = value
	}

	// Copy meta-data from the response to the TemplateData
	for key, value := range rs.meta {
		td.Meta[key] = value
	}

	return td
}

// Write writes the response to the http.ResponseWriter. It implements the ResponseWriter interface.
func (rs *Response) Write(w http.ResponseWriter, r *http.Request) error {
	templateData := rs.ToTemplateData(r)

	// Add all headers to the response, including triggers from htmx
	for k, v := range templateData.Headers {
		w.Header().Set(k, v)
	}

	// Get template
	tmpl, err := rs.tm.Page(rs.path)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrTemplateNotFound) {
			status = http.StatusNotFound
		}
		return rs.WriteError(w, r, status, err)
	}

	// Set status code before writing body
	w.WriteHeader(rs.status)

	// Execute template
	if err := tmpl.ExecuteTemplate(w, fmt.Sprintf("layout:%s", rs.layout), templateData); err != nil {
		return rs.WriteError(w, r, http.StatusInternalServerError, fmt.Errorf("template execution error: %w", err))
	}

	return nil
}

// WriteError renders an error page, falling back to http.Error if that fails
func (rs *Response) WriteError(w http.ResponseWriter, r *http.Request, status int, originalErr error) error {
	errorPath := path.Join(rs.tm.ErrorsDir(), strconv.Itoa(status))
	if errorTmpl, err := rs.tm.Page(errorPath); err == nil {
		rs.Status(status)
		rs.Message(alert.TypeError, originalErr.Error())
		rs.Title(fmt.Sprintf("%d - %s", status, http.StatusText(status)))
		templateData := rs.ToTemplateData(r)
		w.WriteHeader(status)
		if err := errorTmpl.ExecuteTemplate(w, fmt.Sprintf("layout:%s", rs.layout), templateData); err == nil {
			return nil
		}
	}

	// If rendering the error template fails, fall back to http.Error
	http.Error(w, fmt.Sprintf("%d - %s: %v", status, http.StatusText(status), originalErr), status)
	return originalErr
}

// WriteNotFound is a shortcut for writing a 404 Not Found response
func (rs *Response) WriteNotFound(w http.ResponseWriter, r *http.Request) error {
	return rs.WriteError(w, r, http.StatusNotFound, fmt.Errorf("page not found"))
}

// WriteInternalServerError is a shortcut for writing a 500 Internal Server Error response
func (rs *Response) WriteInternalServerError(w http.ResponseWriter, r *http.Request, err error) error {
	return rs.WriteError(w, r, http.StatusInternalServerError, fmt.Errorf("internal server error: %w", err))
}
