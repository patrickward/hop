package conversions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/conversions"
)

func TestToString(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{123, "123"},
		{123.45, "123.45"},
		{true, "true"},
		{"test", "test"},
	}

	for _, tt := range tests {
		result := conversions.FuncMap()["to_string"].(func(any) string)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestToNumber(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
		err      bool
	}{
		{123, 123.0, false},
		{123.45, 123.45, false},
		{"123.45", 123.45, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := conversions.FuncMap()["to_number"].(func(any) (float64, error))(tt.input)
		if tt.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
		err      bool
	}{
		{123, 123, false},
		{123.45, 123, false},
		{"123", 123, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := conversions.FuncMap()["to_int"].(func(any) (int64, error))(tt.input)
		if tt.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestToFloat(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
		err      bool
	}{
		{123, 123.0, false},
		{123.45, 123.45, false},
		{"123.45", 123.45, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := conversions.FuncMap()["to_float"].(func(any) (float64, error))(tt.input)
		if tt.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}
