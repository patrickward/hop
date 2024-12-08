package conf

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	// sensitiveFields defines patterns for field names that should be redacted
	sensitiveFields = []string{
		"password",
		"secret",
		"key",
		"token",
		"credential",
	}

	// maskChar is the character used for masking
	maskChar = "*"
)

// PrettyString returns a formatted string representation of the configuration
func PrettyString(cfg interface{}) string {
	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var sb strings.Builder
	prettyPrint(val, "", &sb)
	return sb.String()
}

func prettyPrint(val reflect.Value, prefix string, sb *strings.Builder) {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		fieldName := fieldType.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		// Handle nested structs (except Duration which is a special case)
		if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(Duration{}) {
			prettyPrint(field, fieldName, sb)
			continue
		}

		// Get the formatted value
		value := formatValue(field, fieldType.Name)
		_, _ = fmt.Fprintf(sb, "%-40s = %s\n", fieldName, value)
	}
}

// formatValue returns the formatted value, masking sensitive data
func formatValue(field reflect.Value, fieldName string) string {
	// Check if this is a sensitive field
	if isSensitive(fieldName) {
		return maskValue(field)
	}

	// Format non-sensitive values
	switch field.Kind() {
	case reflect.String:
		return fmt.Sprintf("%q", field.String())
	case reflect.Bool:
		return fmt.Sprintf("%v", field.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", field.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", field.Float())
	default:
		// Handle Duration type
		if field.Type() == reflect.TypeOf(Duration{}) {
			d := field.Interface().(Duration)
			return fmt.Sprintf("%v", d.Duration)
		}
		return fmt.Sprintf("%v", field.Interface())
	}
}

// isSensitive checks if a field name indicates sensitive data
func isSensitive(fieldName string) bool {
	fieldName = strings.ToLower(fieldName)
	for _, pattern := range sensitiveFields {
		if strings.Contains(fieldName, pattern) {
			return true
		}
	}
	return false
}

// maskValue returns a masked version of the value
func maskValue(field reflect.Value) string {
	switch field.Kind() {
	case reflect.String:
		val := field.String()
		if len(val) == 0 {
			return `""`
		}
		// Show first and last character if string is long enough
		if len(val) > 4 {
			return fmt.Sprintf("%q", val[:1]+strings.Repeat(maskChar, 3)+val[len(val)-1:])
		}
		return fmt.Sprintf("%q", strings.Repeat(maskChar, len(val)))
	default:
		return strings.Repeat(maskChar, 8)
	}
}
