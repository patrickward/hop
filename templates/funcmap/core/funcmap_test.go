package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/core"
)

func TestWhen(t *testing.T) {
	tests := []struct {
		condition bool
		a         any
		expected  any
	}{
		{true, "value", "value"},
		{false, "value", ""},
	}

	for _, tt := range tests {
		result := core.FuncMap()["when"].(func(bool, any) any)(tt.condition, tt.a)
		assert.Equal(t, tt.expected, result)
	}
}

func TestUnless(t *testing.T) {
	tests := []struct {
		condition bool
		a         any
		expected  any
	}{
		{false, "value", "value"},
		{true, "value", ""},
	}

	for _, tt := range tests {
		result := core.FuncMap()["unless"].(func(bool, any) any)(tt.condition, tt.a)
		assert.Equal(t, tt.expected, result)
	}
}

func TestDefaultValue(t *testing.T) {
	tests := []struct {
		value        any
		defaultValue any
		expected     any
	}{
		{"", "default", "default"},
		{"value", "default", "value"},
		{0, 42, 42},
		{10, 42, 10},
	}

	for _, tt := range tests {
		result := core.FuncMap()["default"].(func(any, any) any)(tt.value, tt.defaultValue)
		assert.Equal(t, tt.expected, result)
	}
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		values   []any
		expected any
	}{
		{[]any{"", nil, "value"}, "value"},
		{[]any{nil, "", 0, 42}, 42},
		{[]any{nil, "", 0, false}, nil},
	}

	for _, tt := range tests {
		result := core.FuncMap()["coalesce"].(func(...any) any)(tt.values...)
		assert.Equal(t, tt.expected, result)
	}
}
