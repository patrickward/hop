package conversions

import (
	"fmt"
	"html/template"
	"strconv"
)

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"to_number": toNumber, // Convert any type to a number
		"to_float":  toFloat,  // Convert any type to a float
		"to_int":    toInt,    // Convert any type to an integer
		"to_string": toString, // Convert any type to a string
	}
}

// toString converts any type to a string
func toString(i any) string {
	return fmt.Sprintf("%v", i)
}

func toNumber(i any) (float64, error) {
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
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to number", i)
	}
}

func toInt(i any) (int64, error) {
	switch v := i.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", i)
	}
}

func toFloat(i any) (float64, error) {
	switch v := i.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", i)
	}
}
