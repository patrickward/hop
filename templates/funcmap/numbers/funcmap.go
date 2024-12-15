package numbers

import (
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var printer = message.NewPrinter(language.English)

// FuncMap returns a function map with functions for working with numbers.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"num_add":      add,                // Add two integers
		"num_cents":    cents,              // Format a number as currency with cents
		"num_currency": currency,           // Format a number as currency
		"num_decr":     decrement,          // Decrement a number by 1
		"num_format":   format,             // Format any number with commas and decimals
		"num_incr":     increment,          // Increment a number by 1
		"num_mod":      mod,                // Get the remainder of a division
		"num_percent":  percentage,         // Format a number as a percentage
		"num_sci":      scientificNotation, // Format a number in scientific notation
		"num_sub":      subtract,           // Subtract two integers
	}
}

// add adds two integers
func add(a, b any) (float64, error) {
	// Convert first number
	aFloat, err := toFloat64(a)
	if err != nil {
		return 0, fmt.Errorf("first argument: %w", err)
	}

	// Convert second number
	bFloat, err := toFloat64(b)
	if err != nil {
		return 0, fmt.Errorf("second argument: %w", err)
	}

	return aFloat + bFloat, nil
}

// subtract subtracts two integers
func subtract(a, b any) (float64, error) {
	// Convert first number
	aFloat, err := toFloat64(a)
	if err != nil {
		return 0, fmt.Errorf("first argument: %w", err)
	}

	// Convert second number
	bFloat, err := toFloat64(b)
	if err != nil {
		return 0, fmt.Errorf("second argument: %w", err)
	}

	return aFloat - bFloat, nil
}

// mod returns the remainder of a divided by b
func mod(a, b int) int {
	if b == 0 {
		return 0
	}
	return a % b
}

// Increment an integer value
func increment(i any) (float64, error) {
	val, err := toFloat64(i)
	if err != nil {
		return 0, err
	}
	return val + 1, nil
}

// decrement an integer value
func decrement(i any) (float64, error) {
	val, err := toFloat64(i)
	if err != nil {
		return 0, err
	}
	return val - 1, nil
}

// format formats a number with commas and decimals
// Returns original input as string if not numeric
func format(i any, decimals int) string {
	// For integers: comma separated, ignore decimals
	// For floats: comma separated with decimal places
	// Non-numeric: convert to string

	if decimals < 0 {
		decimals = 0
	}

	switch v := i.(type) {
	case int, int8, int16, int32, int64:
		n, _ := toInt64(v)
		return printer.Sprintf("%d", n) // Uses locale for comma separation
	case float32, float64:
		f, _ := toFloat64(v)
		return printer.Sprintf("%."+strconv.Itoa(decimals)+"f", f)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// cents formats a number as currency with cents by dividing by 100 first
func cents(symbol string, i any) string {
	// Formats as currency with cents
	// Divides by 100 and shows 2 decimal places
	// Returns "$0.00" for non-numeric

	format := symbol + "%.2f"

	switch v := i.(type) {
	case int, int8, int16, int32, int64:
		n, _ := toInt64(v)
		return printer.Sprintf(format, float64(n)/100)
	case float32, float64:
		f, _ := toFloat64(v)
		return printer.Sprintf(format, f/100)
	default:
		return symbol + "0.00"
	}
}

// currency formats a number as currency
func currency(symbol string, decimals int, i any) string {
	// Handles currency formatting with symbol
	// Always shows decimals (default 2)
	// Returns "$0.00" for non-numeric
	if decimals < 0 {
		decimals = 2 // Default for currency
	}

	format := symbol + "%." + strconv.Itoa(decimals) + "f"

	switch v := i.(type) {
	case int, int8, int16, int32, int64:
		n, _ := toInt64(v)
		return printer.Sprintf(format, float64(n))
	case float32, float64:
		f, _ := toFloat64(v)
		return printer.Sprintf(format, f)
	default:
		return symbol + "0." + strings.Repeat("0", decimals)
	}
}

// scientificNotation formats a number in scientific notation
func scientificNotation(i any, decimals int) string {
	// Scientific notation formatting
	// Returns original string if not numeric
	if decimals < 0 {
		decimals = 0
	}

	// Use fmt.Sprintf instead of printer.Sprintf to get standard E notation
	format := "%." + strconv.Itoa(decimals) + "E"

	switch v := i.(type) {
	case int, int8, int16, int32, int64:
		n, _ := toInt64(v)
		return fmt.Sprintf(format, float64(n))
	case float32, float64:
		f, _ := toFloat64(v)
		return fmt.Sprintf(format, f)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// percentage formats a number as a percentage
func percentage(i any, decimals int) string {
	// Formats as percentage
	// Multiplies by 100 and adds %
	// Returns "0%" for non-numeric

	if decimals < 0 {
		decimals = 0
	}

	format := "%." + strconv.Itoa(decimals) + "f%%"

	switch v := i.(type) {
	case int, int8, int16, int32, int64:
		n, _ := toInt64(v)
		return printer.Sprintf(format, float64(n)*100)
	case float32, float64:
		f, _ := toFloat64(v)
		return printer.Sprintf(format, f*100)
	default:
		return "0%"
	}
}

// Helper functions for type conversion
func toInt64(i any) (int64, error) {
	switch v := i.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, fmt.Errorf("uint64 value %d overflows int64", v)
		}
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", i)
	}
}

func toFloat64(i any) (float64, error) {
	switch v := i.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", i)
	}
}
