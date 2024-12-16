package attr

import (
	"fmt"
	"html/template"
	"slices"
	"strings"
)

// FuncMap returns a template.FuncMap for HTML templates
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"attr_class":    classes,                  // Return "class" attribute with the provided classes
		"attr_safe":     safeAttr,                 // Mark a string as safe for HTML attribute output
		"attr_selected": selectedAttr("selected"), // Return "selected" if the value is true, otherwise ""
		"attr_checked":  selectedAttr("checked"),  // Return "checked" if the value is true, otherwise ""
		"attr_disabled": selectedAttr("disabled"), // Return "disabled" if the value is true, otherwise ""
		"attr_readonly": selectedAttr("readonly"), // Return "readonly" if the value is true, otherwise ""
	}
}

// classes returns a template.HTMLAttr value for the "class" attribute
//
// # The function accepts any number of strings or functions that return strings
//
// Example:
//
//	{{ attr_class "card" (when .IsLarge "card-lg") (unless .IsVisible "hidden") }}
func classes(pairs ...any) template.HTMLAttr {
	var classes []string
	for _, pair := range pairs {
		if s, ok := pair.(string); ok {
			// remove any leading/trailing whitespace, quotes, or brackets
			s = strings.Trim(s, " \t\n\r\"'[]")
			if s != "" {
				classes = append(classes, s)
			}
		}
	}

	// return empty string if no classes
	if len(classes) == 0 {
		return ""
	}

	return template.HTMLAttr(fmt.Sprintf(`class="%s"`, strings.Join(classes, " ")))
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
