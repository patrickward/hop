package slices_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/slices"
)

func TestNew(t *testing.T) {
	tests := []struct {
		items    []any
		expected []any
	}{
		{[]any{1, 2, 3}, []any{1, 2, 3}},
		{[]any{"a", "b", "c"}, []any{"a", "b", "c"}},
		{[]any{}, []any{}},
	}

	for _, tt := range tests {
		result := slices.FuncMap()["slc_new"].(func(...any) []any)(tt.items...)
		assert.Equal(t, tt.expected, result)
	}
}

func TestHas(t *testing.T) {
	tests := []struct {
		slice    []any
		item     any
		expected bool
	}{
		{[]any{1, 2, 3}, 2, true},
		{[]any{"a", "b", "c"}, "d", false},
		{[]any{}, 1, false},
	}

	for _, tt := range tests {
		result := slices.FuncMap()["slc_has"].(func([]any, any) bool)(tt.slice, tt.item)
		assert.Equal(t, tt.expected, result)
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		slice    []any
		size     int
		expected [][]any
	}{
		{[]any{1, 2, 3, 4, 5}, 2, [][]any{{1, 2}, {3, 4}, {5}}},
		{[]any{"a", "b", "c", "d"}, 3, [][]any{{"a", "b", "c"}, {"d"}}},
		{[]any{"a", "b", "c"}, 4, [][]any{{"a", "b", "c"}}},
		{[]any{"a", "b", "c"}, 0, nil},
		{[]any{}, 1, [][]any{{}}},
	}

	for _, tt := range tests {
		result := slices.FuncMap()["slc_chunk"].(func([]any, int) [][]any)(tt.slice, tt.size)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGroup(t *testing.T) {
	tests := []struct {
		slice    []any
		keyFn    func(any) any
		expected map[any][]any
	}{
		{
			[]any{1, 2, 3, 4, 5},
			func(v any) any { return v.(int) % 2 },
			map[any][]any{0: {2, 4}, 1: {1, 3, 5}},
		},
		{
			[]any{"apple", "banana", "cherry"},
			func(v any) any { return len(v.(string)) },
			map[any][]any{5: {"apple"}, 6: {"banana", "cherry"}},
		},
	}

	for _, tt := range tests {
		result := slices.FuncMap()["slc_group"].(func([]any, func(any) any) map[any][]any)(tt.slice, tt.keyFn)
		assert.Equal(t, tt.expected, result)
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		slice    []any
		expected []any
	}{
		{[]any{1, 2, 2, 3, 3, 3}, []any{1, 2, 3}},
		{[]any{"a", "b", "a", "c", "b"}, []any{"a", "b", "c"}},
		{[]any{}, []any{}},
	}

	for _, tt := range tests {
		result := slices.FuncMap()["slc_unique"].(func([]any) []any)(tt.slice)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		slice    []any
		expected []any
	}{
		{[]any{3, 1, 2}, []any{1, 2, 3}},
		{[]any{"b", "a", "c"}, []any{"a", "b", "c"}},
		{[]any{2.2, 1.1, 3.3}, []any{1.1, 2.2, 3.3}},
	}

	for _, tt := range tests {
		result := slices.FuncMap()["slc_sort"].(func([]any) []any)(tt.slice)
		assert.Equal(t, tt.expected, result)
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		slice    []any
		expected []any
	}{
		{[]any{1, 2, 3}, []any{3, 2, 1}},
		{[]any{"a", "b", "c"}, []any{"c", "b", "a"}},
		{[]any{}, []any{}},
	}

	for _, tt := range tests {
		result := slices.FuncMap()["slc_reverse"].(func([]any) []any)(tt.slice)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		slice    []any
		pred     func(any) bool
		expected []any
	}{
		{
			[]any{1, 2, 3, 4, 5},
			func(v any) bool { return v.(int) > 3 },
			[]any{4, 5},
		},
		{
			[]any{"apple", "banana", "cherry"},
			func(v any) bool { return len(v.(string)) > 5 },
			[]any{"banana", "cherry"},
		},
	}

	for _, tt := range tests {
		result := slices.FuncMap()["slc_filter"].(func([]any, func(any) bool) []any)(tt.slice, tt.pred)
		assert.Equal(t, tt.expected, result)
	}
}
