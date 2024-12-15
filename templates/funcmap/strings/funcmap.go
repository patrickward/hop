package strings

import (
	"fmt"
	"html/template"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var caser = cases.Title(language.English)

// FuncMap returns a function map with functions for working with strings.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"str_contains":  strings.Contains,  // Check if a string contains a substring
		"str_hasPrefix": strings.HasPrefix, // Check if a string has a prefix
		"str_hasSuffix": strings.HasSuffix, // Check if a string has a suffix
		"str_join":      strings.Join,      // Join a slice into a string
		"str_lower":     strings.ToLower,   // Convert a string to lowercase
		"str_split":     strings.Split,     // Split a string into a slice
		"str_titleize":  Titleize,          // Capitalize the first letter of each word in a string
		"str_toString":  ToString,          // Convert any type to a string
		"str_trim":      strings.Trim,      // Trim a string using a set of characters
		"str_trimSpace": strings.TrimSpace, // Trim leading and trailing spaces from a string
		"str_truncate":  Truncate,          // Truncate a string to a specified length
		"str_upper":     strings.ToUpper,   // Convert a string to uppercase
	}
}

// Titleize capitalizes the first letter of each word in a string (English)
func Titleize(s string) string {
	return caser.String(s)
}

// ToString converts any type to a string
func ToString(i any) string {
	return fmt.Sprintf("%v", i)
}

// Truncate truncates a string to a specified length and appends "..." if longer
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	// Find last space before length
	if idx := strings.LastIndex(s[:length], " "); idx != -1 {
		return s[:idx] + "..."
	}
	return s[:length] + "..."
}
