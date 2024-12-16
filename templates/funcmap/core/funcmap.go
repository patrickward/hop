package core

import (
	"html/template"
	"reflect"
)

// FuncMap returns a template.FuncMap for core functions
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"when":     when,
		"unless":   unless,
		"default":  defaultValue,
		"coalesce": coalesce,
	}
}

// when returns the first value if the condition is true, otherwise the second value
func when(condition bool, a any) any {
	if condition {
		return a
	}

	return ""
}

// unless returns the first value if the condition is false, otherwise the second value
func unless(condition bool, a any) any {
	if !condition {
		return a
	}

	return ""
}

// default returns defaultValue if value is zero/empty, otherwise returns value
func defaultValue(value, defaultValue any) any {
	if isZero(value) {
		return defaultValue
	}
	return value
}

// coalesce returns the first non-zero value in the list
func coalesce(values ...any) any {
	for _, v := range values {
		if !isZero(v) {
			return v
		}
	}
	return nil
}

// isZero checks if a value is empty/zero
func isZero(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() == 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() == 0
	case float32, float64:
		return reflect.ValueOf(v).Float() == 0
	case bool:
		return !v
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	}

	return reflect.ValueOf(value).IsZero()
}
