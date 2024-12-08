package conf

import (
	"fmt"
	"reflect"
	"strings"
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

		// Format the value based on its type
		var value string
		switch field.Kind() {
		case reflect.String:
			value = fmt.Sprintf("%q", field.String())
		case reflect.Bool:
			value = fmt.Sprintf("%v", field.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = fmt.Sprintf("%d", field.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			value = fmt.Sprintf("%d", field.Uint())
		case reflect.Float32, reflect.Float64:
			value = fmt.Sprintf("%.2f", field.Float())
		default:
			// Handle special case for Duration
			if field.Type() == reflect.TypeOf(Duration{}) {
				d := field.Interface().(Duration)
				value = fmt.Sprintf("%v", d.Duration)
			} else {
				value = fmt.Sprintf("%v", field.Interface())
			}
		}

		_, _ = fmt.Fprintf(sb, "%-40s = %s\n", fieldName, value)
	}
}
