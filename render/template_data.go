package render

import (
	"net/http"

	"github.com/patrickward/hop/v2/alert"
	"github.com/patrickward/hop/v2/render/htmx"
)

type TemplateData struct {
	CSRFToken   string            `json:"csrf_token"`
	Environment string            `json:"environment"`
	Status      int               `json:"status"`
	Headers     map[string]string `json:"headers"`
	Layout      string            `json:"layout"`
	Path        string            `json:"path"`
	Title       string            `json:"title"`
	Data        map[string]any    `json:"data"`
	FieldErrors map[string]string `json:"field_errors"`
	Nonce       string            `json:"nonce"`
	Meta        map[string]any    `json:"meta"`
	Flash       alert.Messages    `json:"flash"`
	Messages    alert.Messages    `json:"messages"`
	request     *http.Request
}

// IsProduction checks if the environment is set to "production".
func (td *TemplateData) IsProduction() bool {
	return td.Environment == "production"
}

// IsDevelopment checks if the environment is set to "development".
func (td *TemplateData) IsDevelopment() bool {
	return td.Environment == "development"
}

// IsStaging checks if the environment is set to "staging".
func (td *TemplateData) IsStaging() bool {
	return td.Environment == "staging"
}

// IsTest checks if the environment is set to "test".
func (td *TemplateData) IsTest() bool {
	return td.Environment == "test"
}

// IsHTMXRequest checks if the request is an HTMX request.
func (td *TemplateData) IsHTMXRequest() bool {
	return htmx.IsHtmxRequest(td.request)
}

// IsBoostedRequest checks if the request is a boosted HTMX request.
func (td *TemplateData) IsBoostedRequest() bool {
	return htmx.IsBoostedRequest(td.request)
}

// IsAnyHtmxRequest checks if the request is either an HTMX request or a boosted request.
func (td *TemplateData) IsAnyHtmxRequest() bool {
	return htmx.IsAnyHtmxRequest(td.request)
}

// IsHome checks if the request path is the home path ("/").
func (td *TemplateData) IsHome() bool {
	return td.request != nil && td.request.URL.Path == "/"
}

// RequestPath returns the request path from the TemplateData.
func (td *TemplateData) RequestPath() string {
	if td.request != nil {
		return td.request.URL.Path
	}

	return ""
}

// HasMessages checks if there are any messages in the TemplateData.
func (td *TemplateData) HasMessages() bool {
	return len(td.Messages) > 0
}

// MessagesByType returns messages of a specific type from the TemplateData.
func (td *TemplateData) MessagesByType(msgType alert.Type) alert.Messages {
	return td.Messages.ByType(msgType)
}

// HasFlash checks if there are any flash messages in the TemplateData.
func (td *TemplateData) HasFlash() bool {
	return len(td.Flash) > 0
}

// HasFieldErrors checks if there are any field errors in the TemplateData.
func (td *TemplateData) HasFieldErrors() bool {
	return len(td.FieldErrors) > 0
}

// HasErrorFor checks if there is an error for the specified field.
func (td *TemplateData) HasErrorFor(field string) bool {
	_, ok := td.FieldErrors[field]
	return ok
}

// ErrorFor returns the error message for the specified field.
func (td *TemplateData) ErrorFor(field string) string {
	return td.FieldErrors[field]
}
