package render

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/patrickward/hop/flash"
	"github.com/patrickward/hop/render/htmx"
	"github.com/patrickward/hop/render/htmx/trigger"
)

// Response represents a view response to an HTTP request
// It uses a fluent interface to allow for chaining of methods, so that methods can be called in any order.
type Response struct {
	// The flash manager to be used (default: nil). Flash is only used for rendering flash messages during the render stage.
	flash *flash.Manager
	// The headers to be passed to the response (default: empty)
	headers map[string]string
	// The layout template to be used (required, no default)
	layout string
	// The view template path to be used (required, no default)
	path string
	// The status code to be passed to the response (default: http.StatusOK)
	request *http.Request
	// The status code to be passed to the response (default: http.StatusOK)
	statusCode int
	// The title of the page (default: the page name without the extension)
	title string
	// The description of the page (default: empty)
	description string
	// The triggers to be passed to the response (default: empty)
	triggers *trigger.Triggers
	// The view data to be passed to the template (default: PageData{})
	data *PageData
	// The template manager to be used for rendering templates
	tm *TemplateManager
}

// NewResponse creates a new Response struct with the provided template manager.
func NewResponse(tm *TemplateManager, manager *flash.Manager) *Response {
	return &Response{
		flash:      manager,
		data:       NewPageData(make(map[string]any)),
		headers:    map[string]string{},
		layout:     "",
		path:       "",
		request:    nil,
		statusCode: http.StatusOK,
		title:      "",
		triggers:   trigger.NewTriggers(),
		tm:         tm,
	}
}

// PageData returns the page data model. The request is set here to ensure
// the request is available in the template and that it is not overwritten until later in the process.
func (resp *Response) PageData(r *http.Request) *PageData {
	resp.data.SetTitle(resp.title)
	resp.data.SetRequest(r)
	resp.addFlash(r)
	return resp.data
}

// addFlash adds flash messages to the view data model
func (resp *Response) addFlash(r *http.Request) {
	if resp.flash == nil {
		return
	}

	if flashMessages := resp.flash.Get(r.Context()); len(flashMessages) > 0 {
		fmt.Printf("FOUND FALSH MESSAGES %v\n", flashMessages)
		resp.data.Set("Flash", flashMessages)
	} else {
		fmt.Printf("NO FLASHM MESSASGE\n")
		resp.data.Set("Flash", nil)
	}
}

// GetHeaders returns the headers map as a combination map of both triggers and headers
func (resp *Response) GetHeaders() map[string]string {
	if resp.headers == nil {
		resp.headers = map[string]string{}
	}

	if resp.triggers != nil {
		if resp.triggers.HasTriggers() {
			val, err := resp.triggers.TriggerHeader()
			if err == nil {
				resp.headers[htmx.HXTrigger] = val
			}
		}

		if resp.triggers.HasAfterSwapTriggers() {
			val, err := resp.triggers.TriggerAfterSwapHeader()
			if err == nil {
				resp.headers[htmx.HXTriggerAfterSwap] = val
			}
		}

		if resp.triggers.HasAfterSettleTriggers() {
			val, err := resp.triggers.TriggerAfterSettleHeader()
			if err == nil {
				resp.headers[htmx.HXTriggerAfterSettle] = val
			}
		}
	}

	return resp.headers
}

// GetHTTPHeader returns a http.Header for the headers map
func (resp *Response) GetHTTPHeader() http.Header {
	if resp.headers == nil {
		return nil
	}

	header := make(http.Header)
	for key, value := range resp.GetHeaders() {
		header.Set(key, value)
	}

	return header
}

// GetTemplateLayout returns the template layout
func (resp *Response) GetTemplateLayout() string {
	return resp.layout
}

// GetTemplatePath returns the path used in templates, if any
func (resp *Response) GetTemplatePath() string {
	return resp.path
}

// GetPageTitle returns the page title
func (resp *Response) GetPageTitle() string {
	return resp.title
}

// GetPageDescription returns the page description
func (resp *Response) GetPageDescription() string {
	return resp.description
}

// GetStatusCode returns the status code.
func (resp *Response) GetStatusCode() int {
	return resp.statusCode
}

// WithPageData resets the view data model with an existing model. It returns the modified Response pointer.
// The view data model contains data that will be passed to the view template for rendering.
//
// Alternatively, you can add the data map and create a new view data model automatically using the WithData function.
//
// Important: if you are creating the view data model externally and need to use it before render is called,
// you should probably set the request via PageData.SetRequest, as the request is deliberately set later in
// the rendering process in most cases.
func (resp *Response) WithPageData(data *PageData) *Response {
	resp.data = data
	return resp
}

// WithData creates a new view data model with the provided data map and returns the modified Response pointer.
// This will overwrite any existing view data model. If you want to add data to an existing view data model, create
// a new view data model externally using the NewPageData function and pass it to the WithPageData function instead.
func (resp *Response) WithData(data map[string]any) *Response {
	resp.data = NewPageData(data)
	return resp
}

// MergeData adds data to the view data model. It returns the modified Response pointer.
func (resp *Response) MergeData(data map[string]any) *Response {
	resp.data.Merge(data)
	return resp
}

// Data adds a single data item to the view data model. It returns the modified Response pointer.
func (resp *Response) Data(key string, value any) *Response {
	resp.data.Set(key, value)
	return resp
}

// WithError adds an error message to the view data model. This will also set the status code to 422 (Unprocessable Entity)).
func (resp *Response) WithError(msg string) *Response {
	resp.statusCode = http.StatusUnprocessableEntity
	resp.data.Set(PageDataErrorKey, msg)
	return resp
}

// WithErrors adds an error message and a map of field errors to the view data model.
// This will also set the status code to 422 (Unprocessable Entity)).
// If that is not the desired status code, you should reset it using the Status() function or one of the Status* shortcut functions.
func (resp *Response) WithErrors(msg string, fieldErrors map[string]string) *Response {
	resp.statusCode = http.StatusUnprocessableEntity
	resp.data.Set(PageDataErrorKey, msg)
	resp.data.Set(PageDataErrorsKey, fieldErrors)
	return resp
}

// Title sets the page title
func (resp *Response) Title(title string) *Response {
	resp.title = title
	return resp
}

// Description sets the page description
func (resp *Response) Description(description string) *Response {
	resp.description = description
	return resp
}

// Path sets the template path
func (resp *Response) Path(path string) *Response {
	// If the path contains a colon, it's part of a plugin path, so we need to
	// extract the plugin name from the path first
	pathParts := strings.SplitN(path, ":", 2)

	if len(pathParts) == 2 {
		path = pathParts[1]
	}

	if !strings.HasPrefix(path, ViewsDir+"/") {
		path = ViewsDir + "/" + path
	}

	if len(pathParts) == 2 {
		path = pathParts[0] + ":" + path
	}

	resp.path = path
	return resp
}

// Layout sets the template layout. It updates the layout value in the Response struct.
// Then it returns the updated Response struct itself for method chaining.
func (resp *Response) Layout(layout string) *Response {
	resp.layout = layout
	return resp
}

// Header adds/sets a header
func (resp *Response) Header(key, value string) *Response {
	if resp.headers == nil {
		resp.headers = make(map[string]string)
	}

	resp.headers[key] = value
	return resp
}

// Status sets the status code.
func (resp *Response) Status(status int) *Response {
	resp.statusCode = status
	return resp
}

// NoCacheStrict sets the Cache-Control header to "no-cache, no-store, must-revalidate".
func (resp *Response) NoCacheStrict() *Response {
	resp.headers["Cache-Control"] = "no-cache, no-store, must-revalidate"
	return resp
}

// CacheControl sets the Cache-Control header to the given value.
func (resp *Response) CacheControl(cacheControl string) *Response {
	resp.headers["Cache-Control"] = cacheControl
	return resp
}

// ETag sets the ETag header to the given value.
func (resp *Response) ETag(etag string) *Response {
	resp.headers["ETag"] = etag
	return resp
}

// LastModified sets the Last-Modified header to the given value.
func (resp *Response) LastModified(lastModified string) *Response {
	resp.headers["Last-Modified"] = lastModified
	return resp
}
