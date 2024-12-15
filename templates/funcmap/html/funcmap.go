package html

import (
	"fmt"
	"html/template"
	"net/url"
	"slices"
)

// FuncMap returns a template.FuncMap for HTML templates
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"htm_attr":        safeAttr,                 // Mark a string as safe for HTML attribute output
		"htm_safe":        safeHTML,                 // Mark a string as safe for HTML output
		"htm_urlSetParam": urlSetParam,              // Set a query parameter in a URL
		"htm_urlDelParam": urlDelParam,              // Delete a query parameter from a URL
		"htm_urlToAttr":   urlToAttr,                // Convert a URL to a template.HTMLAttr value
		"htm_selected":    selectedAttr("selected"), // Return "selected" if the value is true, otherwise ""
		"htm_checked":     selectedAttr("checked"),  // Return "checked" if the value is true, otherwise ""
		"htm_disabled":    selectedAttr("disabled"), // Return "disabled" if the value is true, otherwise ""
		"htm_readonly":    selectedAttr("readonly"), // Return "readonly" if the value is true, otherwise ""
	}
}

// safeHTML returns a template.HTML value to mark a string as safe for HTML output
func safeHTML(s string) template.HTML {
	return template.HTML(s)
}

// safeAttr returns a template.HTMLAttr value to mark a string as safe for HTML attribute output
func safeAttr(s string) template.HTMLAttr {
	return template.HTMLAttr(s)
}

// selectedAttr returns a function that returns the provided attribute if the value is equal to the current value
func selectedAttr(attr string) func(current any, value string) template.HTMLAttr {
	return func(current any, value string) template.HTMLAttr {
		// Switch on the type of current value
		// We can accept strings or slices of strings
		switch c := current.(type) {
		case string:
			if value == c {
				return template.HTMLAttr(attr)
			}
		case []string:
			if slices.Contains(c, value) {
				return template.HTMLAttr(attr)
			}
		}

		// convert current to string and compare
		if fmt.Sprintf("%v", current) == value {
			return template.HTMLAttr(attr)
		}

		return ""
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
