package slices

import (
	"fmt"
	"html/template"
	"reflect"
	"sort"
)

// FuncMap returns slice-specific functions
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"slc_chunk":   Chunk,   // Was: "group"
		"slc_filter":  Filter,  // Filter slice by predicate
		"slc_group":   Group,   // Was: N/A"
		"slc_has":     Has,     // Was: "has"
		"slc_new":     New,     // Was: "slice" - Creates new slice
		"slc_reverse": Reverse, // Reverse sort a slice
		"slc_sort":    Sort,    // Sort any slice
		"slc_unique":  Unique,  // Get unique elements
	}
}

// New creates a new slice from the provided items
//
// Example: {{ slice.new 1 2 3 4 5 }}
func New(items ...any) []any {
	return items
}

// Has checks if a slice contains an item
//
// Example: {{ slice.has .Slice "item" }}
func Has(s []any, item any) bool {
	for _, v := range s {
		if reflect.DeepEqual(v, item) {
			return true
		}
	}
	return false
}

// Chunk splits a slice into chunks of the specified size
//
// Example: {{ slice.chunk .Slice 3 }} -> [[1 2 3] [4 5 6] [7 8 9]]
func Chunk(items []any, size int) [][]any {
	if size <= 0 {
		return nil
	}

	chunks := make([][]any, 0, (len(items)+size-1)/size)
	for size < len(items) {
		items, chunks = items[size:], append(chunks, items[0:size:size])
	}
	chunks = append(chunks, items)
	return chunks
}

// Group groups slice elements by a key function
//
// Example: {{ slice.group .Items (index . "key") }} -> map[key:[item1 item2] key2:[item3 item4]]
func Group(items []any, keyFn func(any) any) map[any][]any {
	groups := make(map[any][]any)
	for _, item := range items {
		key := keyFn(item)
		groups[key] = append(groups[key], item)
	}
	return groups
}

// Unique returns unique elements from a slice
//
// Example: {{ slice.unique .Slice }} -> [1 2 3 4 5]
func Unique(items []any) []any {
	seen := make(map[any]struct{})
	result := make([]any, 0, len(items))

	for _, item := range items {
		// Use string representation for maps/slices as they can't be map keys
		var key any
		if reflect.TypeOf(item) == nil {
			key = nil
		} else if reflect.TypeOf(item).Kind() == reflect.Map ||
			reflect.TypeOf(item).Kind() == reflect.Slice {
			key = fmt.Sprintf("%#v", item)
		} else {
			key = item
		}

		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// Sort sorts a slice in ascending order
//
// Example: {{ slice.sort .Slice }} -> [1 2 3 4 5]
func Sort(items []any) []any {
	result := make([]any, len(items))
	copy(result, items)

	// If empty or nil slice, return as is
	if len(result) == 0 {
		return result
	}

	// Check first element type and sort accordingly
	switch result[0].(type) {
	case string:
		strings := make([]string, len(result))
		for i, v := range result {
			strings[i] = v.(string)
		}
		sort.Strings(strings)
		for i, v := range strings {
			result[i] = v
		}
	case int:
		ints := make([]int, len(result))
		for i, v := range result {
			ints[i] = v.(int)
		}
		sort.Ints(ints)
		for i, v := range ints {
			result[i] = v
		}
	case float64:
		floats := make([]float64, len(result))
		for i, v := range result {
			floats[i] = v.(float64)
		}
		sort.Float64s(floats)
		for i, v := range floats {
			result[i] = v
		}
	}

	return result
}

// Reverse reverses the elements in a slice
//
// Example: {{ slice.reverse .Slice }} -> [5 4 3 2 1]
func Reverse(items []any) []any {
	result := make([]any, len(items))
	copy(result, items)

	for i := 0; i < len(result)/2; i++ {
		j := len(result) - i - 1
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// Filter returns a new slice containing only the elements that satisfy the predicate
//
// Example: {{ slice.filter .Slice (lambda (x) (gt x 5)) }} -> [6 7 8 9]
func Filter(items []any, pred func(any) bool) []any {
	result := make([]any, 0, len(items))
	for _, item := range items {
		if pred(item) {
			result = append(result, item)
		}
	}
	return result
}
