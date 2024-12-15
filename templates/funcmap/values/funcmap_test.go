package values_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/values"
)

func TestCoalesce(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"", "", "first", "second"}, "first"},
		{[]string{"", "first", "second"}, "first"},
		{[]string{"first", "second"}, "first"},
		{[]string{"", ""}, ""},
	}

	for _, tt := range tests {
		result := values.FuncMap()["val_coalesce"].(func(...string) string)(tt.input...)
		assert.Equal(t, tt.expected, result)
	}
}

func TestYesNo(t *testing.T) {
	tests := []struct {
		input    bool
		expected string
	}{
		{true, "Yes"},
		{false, "No"},
	}

	for _, tt := range tests {
		result := values.FuncMap()["val_yesno"].(func(bool) string)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestOnOff(t *testing.T) {
	tests := []struct {
		input    bool
		expected string
	}{
		{true, "On"},
		{false, "Off"},
	}

	for _, tt := range tests {
		result := values.FuncMap()["val_onoff"].(func(bool) string)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
