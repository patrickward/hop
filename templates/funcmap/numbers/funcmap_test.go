package numbers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/numbers"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		a, b     any
		expected float64
	}{
		{1, 2, 3},
		{-1, -2, -3},
		{0, 0, 0},
		{1.1, 2.2, 3.3},
		{-1.1, -2.2, -3.3},
		{0.0, 0.0, 0.0},
		{1, 2.2, 3.2},
	}

	for _, tt := range tests {
		result, err := numbers.FuncMap()["num_add"].(func(any, any) (float64, error))(tt.a, tt.b)
		assert.NoError(t, err)
		assert.InDelta(t, tt.expected, result, 0.0001)
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		a, b     any
		expected float64
	}{
		{1, 2, -1},
		{-1, -2, 1},
		{0, 0, 0},
		{1.1, 2.2, -1.1},
		{-1.1, -2.2, 1.1},
		{0.0, 0.0, 0.0},
		{1, 2.2, -1.2},
	}

	for _, tt := range tests {
		result, err := numbers.FuncMap()["num_sub"].(func(any, any) (float64, error))(tt.a, tt.b)
		assert.NoError(t, err)
		assert.InDelta(t, tt.expected, result, 0.0001)
	}
}

func TestMod(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{5, 3, 2},
		{5, 0, 0},
		{0, 5, 0},
		{5, -3, 2},
		{-5, 3, -2},
		{-5, -3, -2},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, numbers.FuncMap()["num_mod"].(func(int, int) int)(tt.a, tt.b))
	}
}

func TestDecr(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
	}{
		{1, 0},
		{0, -1},
		{"2", 1},
	}

	for _, tt := range tests {
		result, err := numbers.FuncMap()["num_decr"].(func(any) (float64, error))(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIncr(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
	}{
		{1, 2},
		{0, 1},
		{"2", 3},
	}

	for _, tt := range tests {
		result, err := numbers.FuncMap()["num_incr"].(func(any) (float64, error))(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input    float64
		dp       int
		expected string
	}{
		{1234, 0, "1,234"},
		{1234.5678, 2, "1,234.57"},
		{1234567.89, 2, "1,234,567.89"},
		{1.2345, 2, "1.23"},
		{1.2349, 3, "1.235"},
		{1.0, 2, "1.00"},
		{0.0, 2, "0.00"},
		{0.1234, 2, "0.12"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, numbers.FuncMap()["num_format"].(func(any, int) string)(tt.input, tt.dp))
	}
}

func TestPercentage(t *testing.T) {
	tests := []struct {
		input    float64
		dp       int
		expected string
	}{
		{0.1234, 0, "12%"},
		{0.1234, 1, "12.3%"},
		{0.1234, 2, "12.34%"},
		{0.1234, 3, "12.340%"},
		{0.1234, 4, "12.3400%"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, numbers.FuncMap()["num_percent"].(func(any, int) string)(tt.input, tt.dp))
	}
}

func TestScientificNotation(t *testing.T) {
	tests := []struct {
		input    float64
		dp       int
		expected string
	}{
		{1234, 0, "1E+03"},
		{1234.5678, 2, "1.23E+03"},
		{1234567.89, 2, "1.23E+06"},
		{1.2345, 2, "1.23E+00"},
		{1.2349, 3, "1.235E+00"},
		{1.0, 2, "1.00E+00"},
		{0.0, 2, "0.00E+00"},
		{0.1234, 2, "1.23E-01"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, numbers.FuncMap()["num_sci"].(func(any, int) string)(tt.input, tt.dp))
	}
}

func TestCents(t *testing.T) {
	tests := []struct {
		input    any
		symbol   string
		expected string
	}{
		{1234, "$", "$12.34"},
		{12345, "£", "£123.45"},
		{123456, "€", "€1,234.56"},
		{1234567, "¥", "¥12,345.67"},
		{12345678, "₹", "₹123,456.78"},
		{123456789, "₩", "₩1,234,567.89"},
		{1234567890, "₪", "₪12,345,678.90"},
		{12345678901, "₫", "₫123,456,789.01"},
		{1, "$", "$0.01"},
		{0, "$", "$0.00"},
		{0.1, "$", "$0.00"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, numbers.FuncMap()["num_cents"].(func(string, any) string)(tt.symbol, tt.input))
	}
}

func TestCurrency(t *testing.T) {
	tests := []struct {
		input    float64
		symbol   string
		dp       int
		expected string
	}{
		{1234, "$", 0, "$1,234"},
		{1234.5678, "£", 2, "£1,234.57"},
		{1234567.89, "€", 2, "€1,234,567.89"},
		{1.2345, "¥", 2, "¥1.23"},
		{1.2349, "₹", 3, "₹1.235"},
		{1.0, "₩", 2, "₩1.00"},
		{0.0, "₪", 2, "₪0.00"},
		{0.1234, "₫", 2, "₫0.12"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, numbers.FuncMap()["num_currency"].(func(string, int, any) string)(tt.symbol, tt.dp, tt.input))
	}
}
