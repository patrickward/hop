package collections

import (
	"fmt"
	"html/template"
	"reflect"
	"strings"
)

// FuncMap returns collection-related template functions
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"col_first":    first,    // Get the first element of a collection
		"col_last":     last,     // Get the last element of a collection
		"col_nth":      nth,      // Get the nth element of a collection
		"col_join":     fmtJoin,  // Join collection elements with separator
		"col_list":     fmtList,  // Format collection as a list
		"col_empty":    isEmpty,  // Check if collection is empty
		"col_size":     size,     // Get collection size
		"col_contains": contains, // Check if collection contains value
	}
}

// first returns the first element of a collection
func first(items any) any {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice || v.Len() == 0 {
		return nil
	}
	return v.Index(0).Interface()
}

// last returns the last element of a collection
func last(items any) any {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice || v.Len() == 0 {
		return nil
	}
	return v.Index(v.Len() - 1).Interface()
}

// nth returns the nth element of a collection
func nth(items any, n int) any {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice || n < 0 || n >= v.Len() {
		return nil
	}
	return v.Index(n).Interface()
}

// fmtJoin joins collection elements with separator
func fmtJoin(items any, sep string) string {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return ""
	}

	vals := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		vals[i] = fmt.Sprint(v.Index(i).Interface())
	}
	return strings.Join(vals, sep)
}

// fmtList formats a collection as a list with Oxford comma
func fmtList(items any, sep string, lastSep string) string {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return ""
	}

	switch v.Len() {
	case 0:
		return ""
	case 1:
		return fmt.Sprint(v.Index(0).Interface())
	case 2:
		return fmt.Sprintf("%v%s%v",
			v.Index(0).Interface(),
			lastSep,
			v.Index(1).Interface())
	default:
		vals := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			vals[i] = fmt.Sprint(v.Index(i).Interface())
		}
		last := vals[len(vals)-1]
		rest := strings.Join(vals[:len(vals)-1], sep)
		return rest + lastSep + last
	}
}

// isEmpty checks if collection is empty
func isEmpty(items any) bool {
	v := reflect.ValueOf(items)
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	default:
		return true
	}
}

// size returns collection length
func size(items any) int {
	v := reflect.ValueOf(items)
	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len()
	default:
		return 0
	}
}

// contains checks if collection contains value
func contains(items any, val any) bool {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return false
	}
	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(v.Index(i).Interface(), val) {
			return true
		}
	}
	return false
}
