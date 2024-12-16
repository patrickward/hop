package html

import (
	"html/template"
)

// FuncMap returns a template.FuncMap for HTML templates
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"html_safe": safeHTML, // Mark a string as safe for HTML output
	}
}

// safeHTML returns a template.HTML value to mark a string as safe for HTML output
func safeHTML(s string) template.HTML {
	return template.HTML(s)
}
