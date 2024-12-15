package collections_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/collections"
)

func TestFirst(t *testing.T) {
	tests := []struct {
		input    any
		expected any
	}{
		{[]int{1, 2, 3}, 1},
		{[]string{"a", "b", "c"}, "a"},
		{[]int{}, nil},
	}

	for _, tt := range tests {
		result := collections.FuncMap()["col_first"].(func(any) any)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestLast(t *testing.T) {
	tests := []struct {
		input    any
		expected any
	}{
		{[]int{1, 2, 3}, 3},
		{[]string{"a", "b", "c"}, "c"},
		{[]int{}, nil},
	}

	for _, tt := range tests {
		result := collections.FuncMap()["col_last"].(func(any) any)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestNth(t *testing.T) {
	tests := []struct {
		input    any
		n        int
		expected any
	}{
		{[]int{1, 2, 3}, 1, 2},
		{[]string{"a", "b", "c"}, 2, "c"},
		{[]int{}, 0, nil},
		{[]int{1, 2, 3}, -1, nil},
		{[]int{1, 2, 3}, 3, nil},
	}

	for _, tt := range tests {
		result := collections.FuncMap()["col_nth"].(func(any, int) any)(tt.input, tt.n)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFmtJoin(t *testing.T) {
	tests := []struct {
		input    any
		sep      string
		expected string
	}{
		{[]int{1, 2, 3}, ", ", "1, 2, 3"},
		{[]string{"a", "b", "c"}, "-", "a-b-c"},
		{[]int{}, ", ", ""},
	}

	for _, tt := range tests {
		result := collections.FuncMap()["col_join"].(func(any, string) string)(tt.input, tt.sep)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFmtList(t *testing.T) {
	tests := []struct {
		input    any
		sep      string
		lastSep  string
		expected string
	}{
		{[]int{1, 2, 3}, ", ", " and ", "1, 2 and 3"},
		{[]string{"a", "b", "c"}, ", ", " or ", "a, b or c"},
		{[]int{1}, ", ", " and ", "1"},
		{[]int{}, ", ", " and ", ""},
	}

	for _, tt := range tests {
		result := collections.FuncMap()["col_list"].(func(any, string, string) string)(tt.input, tt.sep, tt.lastSep)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		input    any
		expected bool
	}{
		{[]int{1, 2, 3}, false},
		{[]string{}, true},
		{map[string]int{}, true},
		{map[string]int{"a": 1}, false},
	}

	for _, tt := range tests {
		result := collections.FuncMap()["col_empty"].(func(any) bool)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		input    any
		expected int
	}{
		{[]int{1, 2, 3}, 3},
		{[]string{}, 0},
		{map[string]int{}, 0},
		{map[string]int{"a": 1}, 1},
	}

	for _, tt := range tests {
		result := collections.FuncMap()["col_size"].(func(any) int)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		input    any
		val      any
		expected bool
	}{
		{[]int{1, 2, 3}, 2, true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]int{}, 1, false},
	}

	for _, tt := range tests {
		result := collections.FuncMap()["col_contains"].(func(any, any) bool)(tt.input, tt.val)
		assert.Equal(t, tt.expected, result)
	}
}
