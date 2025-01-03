package render

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
)

// Render renders the response using the template manager
// Example: resp.StatusOK().Render(w, r)
func (resp *Response) Render(w http.ResponseWriter, r *http.Request) {
	// Enforce a layout if none is set
	if resp.GetTemplateLayout() == "" {
		resp.Layout(resp.tm.baseLayout)
	}
	resp.tm.render(w, r, resp)
}

// RenderUnauthorized renders the 401 Unauthorized page
func (resp *Response) RenderUnauthorized(w http.ResponseWriter, r *http.Request) {
	resp.tm.renderSystemError(w, r, resp, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
}

// RenderForbidden renders the 403 Forbidden page
func (resp *Response) RenderForbidden(w http.ResponseWriter, r *http.Request) {
	resp.tm.renderSystemError(w, r, resp, http.StatusForbidden, fmt.Errorf("forbidden"))
}

// RenderMethodNotAllowed renders the 405 Method Not Allowed page
func (resp *Response) RenderMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	resp.tm.renderSystemError(w, r, resp, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
}

// RenderNotFound renders the 404 Not Found page
func (resp *Response) RenderNotFound(w http.ResponseWriter, r *http.Request) {
	resp.tm.renderSystemError(w, r, resp, http.StatusNotFound, fmt.Errorf("not found"))
}

// RenderMaintenance renders the 503 Service Unavailable page
func (resp *Response) RenderMaintenance(w http.ResponseWriter, r *http.Request) {
	resp.tm.renderSystemError(w, r, resp, http.StatusServiceUnavailable, fmt.Errorf("service Unavailable"))
}

// RenderSystemError renders the 500 Internal Server Error page
func (resp *Response) RenderSystemError(w http.ResponseWriter, r *http.Request, err error) {
	// Get the stack trace and output to the log
	if resp.tm.logger != nil {
		resp.tm.logger.Error("Server error", slog.String("err", err.Error()))
	}
	lineErrors := ""
	lines := strings.Split(string(debug.Stack()), "\n")
	for i, line := range lines {
		// replace \t with 4 spaces
		line = strings.ReplaceAll(line, "\t", "    ")
		lineErrors += fmt.Sprintf("--- traceLine%03d: %s\n", i, line)
		if resp.tm.logger != nil {
			resp.tm.logger.Error("Stack trace", slog.String(fmt.Sprintf("--- traceLine%03d", i), line))
		}
	}

	// If there is a template with the name "system/server_error" in the template cache, use it
	resp.tm.renderSystemError(w, r, resp, http.StatusInternalServerError, fmt.Errorf("internal server error: %s\n%s", err.Error(), lineErrors))
}
