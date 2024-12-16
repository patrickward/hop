package url

import (
	"fmt"
	"html/template"
	"net/url"
)

// FuncMap returns a template.FuncMap for HTML templates
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"url_set":     urlSetParam,     // Set a query parameter in a URL
		"url_del":     urlDelParam,     // Delete a query parameter from a URL
		"url_to_attr": urlToAttr,       // Convert a URL to a template.HTMLAttr value
		"url_escape":  url.QueryEscape, // Escape a string for use in a URL
	}
}

// urlSetParam sets a query parameter in a URL
func urlSetParam(key string, value any, u *url.URL) *url.URL {
	nu := *u
	values := nu.Query()

	values.Set(key, fmt.Sprintf("%v", value))

	nu.RawQuery = values.Encode()
	return &nu
}

// urlDelParam deletes a query parameter from a URL
func urlDelParam(key string, u *url.URL) *url.URL {
	nu := *u
	values := nu.Query()

	values.Del(key)

	nu.RawQuery = values.Encode()
	return &nu
}

// urlToAttr converts a URL to a template.HTMLAttr value
func urlToAttr(u *url.URL) template.HTMLAttr {
	return template.HTMLAttr(u.String())
}
