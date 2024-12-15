package debug_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/debug"
)

func TestDump(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{123, "123"},
		{[]int{1, 2, 3}, "[\n  1,\n  2,\n  3\n]"},
		{map[string]int{"a": 1}, "{\n  \"a\": 1\n}"},
	}

	for _, tt := range tests {
		result := debug.FuncMap()["dbg_dump"].(func(any) string)(tt.input)
		assert.JSONEq(t, tt.expected, result)
	}
}

func TestTypeof(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{123, "int"},
		{123.45, "float64"},
		{"test", "string"},
		{true, "bool"},
		{[]int{1, 2, 3}, "[]int"},
	}

	for _, tt := range tests {
		result := debug.FuncMap()["dbg_typeof"].(func(any) string)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
