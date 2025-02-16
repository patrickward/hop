package render

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/patrickward/hop/render/request"
)

const (
	PageDataPageKey   = "Page"
	PageDataErrorKey  = "SetError"
	PageDataErrorsKey = "Errors"
)

// PageData is the struct that all view models must implement. It provides common data for all templates
// and represents the data that is passed to the template.
//
// This is a short-lived object that is used to work with data passed to the template. It is not thread-safe.
//
// PageData should not be used directly. Instead, use the NewPageData function to create an instance
// of PageData that contains the data you want to pass to the template.
//
// Example: NewPageData(request, map[string]any{"title": "Hello World"})
//
// The environment variables are generally added via conf/config, but if you're not using that package,
// you can set them manually.
//
//goland:noinspection GoNameStartsWithPackageName
type PageData struct {
	title   string
	request *http.Request
	data    map[string]any
}

// NewPageData creates a new PageData instance.
// If you are using this outside the normal HyperView rendering process, be sure to set the request manually
// via PageData.SetRequest as the request is deliberately set later in the normal rendering flow.
func NewPageData(pageData map[string]any) *PageData {
	pageData = initData(pageData)
	return &PageData{
		data: pageData,
	}
}

// SetTitle sets the title of the page.
func (v *PageData) SetTitle(title string) {
	v.title = title
}

// SetRequest sets the request for the PageData instance.
func (v *PageData) SetRequest(r *http.Request) {
	v.request = r
}

func initData(data map[string]any) map[string]any {
	if data == nil {
		data = map[string]any{}
	}

	// If no "SetError" key is set, set it to an empty string
	if _, ok := data[PageDataErrorKey]; !ok {
		data[PageDataErrorKey] = ""
	}

	// If no "Errors" key is set, set it to an empty map
	if _, ok := data[PageDataErrorsKey]; !ok {
		data[PageDataErrorsKey] = map[string]string{}
	}

	return data
}

// Data returns the data map that will be passed to the template.
// It will include the PageData instance itself as the "Page" key.
func (v *PageData) Data() map[string]any {
	v.data = initData(v.data)
	v.data[PageDataPageKey] = v
	return v.data
}

// Merge adds a map of data to the existing view data model.
func (v *PageData) Merge(data map[string]any) {
	for key, value := range data {
		v.data[key] = value
	}
}

// Set adds a single key-value pair to the existing view data model.
func (v *PageData) Set(key string, value any) {
	v.data[key] = value
}

// Get returns the value of the specified key from the view data model.
func (v *PageData) Get(key string) any {
	val, ok := v.data[key]
	if ok {
		return val
	}

	return ""
}

// GetString returns the value of the specified key from the view data model as a string.
func (v *PageData) GetString(key string) string {
	val, ok := v.Get(key).(string)
	if ok {
		return val
	}

	return ""
}

// Title returns the title of the page.
func (v *PageData) Title() string {
	return v.title
}

// ------ SetError Helpers --------

// Error returns the error message from the view data model.
func (v *PageData) Error() string {
	return v.GetString(PageDataErrorKey)
}

// HasError returns true if the view data model contains an error message.
func (v *PageData) HasError() bool {
	return v.GetString(PageDataErrorKey) != ""
}

// ------ Field SetError Helpers --------

// Errors returns a map of field errors from the view data model.
func (v *PageData) Errors() map[string]string {
	val, ok := v.Get(PageDataErrorsKey).(map[string]string)
	if ok {
		return val
	}

	return map[string]string{}
}

// HasErrors returns true if the view data model contains field errors.
func (v *PageData) HasErrors() bool {
	return len(v.Errors()) > 0
}

// ErrorFor returns the error message for the specified field from the view data model.
func (v *PageData) ErrorFor(field string) string {
	if errs := v.Errors(); len(errs) > 0 {
		if msg, ok := errs[field]; ok {
			return msg
		}
	}

	return ""
}

// HasErrorFor returns true if the view data model contains an error message for the specified field.
func (v *PageData) HasErrorFor(field string) bool {
	return v.ErrorFor(field) != ""
}

// ------ Path Helpers --------

type LinkData struct {
	Path   string
	Title  string
	Active bool
}

// ActiveLink returns a LinkData struct with the provided path and title.
func (v *PageData) ActiveLink(path, title string) LinkData {
	return LinkData{
		Path:   path,
		Title:  title,
		Active: v.RequestPath() == path,
	}
}

// ActivePrefixLink returns a LinkData struct with the provided path and title, where the active state is determined by the prefix.
func (v *PageData) ActivePrefixLink(path, title, prefix string) LinkData {
	return LinkData{
		Path:   path,
		Title:  title,
		Active: strings.HasPrefix(v.RequestPath(), prefix),
	}
}

// ActiveSuffixLink returns a LinkData struct with the provided path and title, where the active state is determined by the suffix.
func (v *PageData) ActiveSuffixLink(path, title, suffix string) LinkData {
	return LinkData{
		Path:   path,
		Title:  title,
		Active: strings.HasSuffix(v.RequestPath(), suffix),
	}
}

// RequestPath returns the path of the request.
func (v *PageData) RequestPath() string {
	return request.URLPath(v.request)
}

// RequestMethod returns the method of the request.
func (v *PageData) RequestMethod() string {
	return request.Method(v.request)
}

// ------ Common Helpers --------

// BaseURL returns the base URL of the request.
func (v *PageData) BaseURL() string {
	return request.BaseURL(v.request)
}

// Context returns the context of the request.
func (v *PageData) Context() context.Context {
	return v.request.Context()
}

// CurrentYear returns the current year.
func (v *PageData) CurrentYear() int {
	return time.Now().Year()
}

// Nonce returns the nonce value from the request context, if available.
func (v *PageData) Nonce() string {
	nonce, ok := v.request.Context().Value(NonceContextKey).(string)
	if ok {
		return nonce
	}

	return ""
}
