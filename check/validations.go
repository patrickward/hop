package check

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	rgxEmail = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	rgxPhone = `^\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})$`
)

// Required checks if a value is non-empty
//
// Example usage:
// Required("some value") // returns true
// Required("") // returns false
// Required([]int{1, 2, 3}) // returns true
// Required([]int{}) // returns false
func Required(value any) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	case time.Time:
		return !v.IsZero()
	default:
		return true
	}
}

// NotZero checks if a numeric value is not a zero value
func NotZero[T comparable](value T) bool {
	return value != *new(T)
}

// MinLength returns a validation function that checks minimum string length
//
// Example usage:
// MinLength(5)("hello") // returns true
// MinLength(5)("hi") // returns false
func MinLength(min int) ValidationFunc {
	return func(value any) bool {
		str, ok := value.(string)
		if !ok {
			return false
		}
		return utf8.RuneCountInString(strings.TrimSpace(str)) >= min
	}
}

// MaxLength returns a validation function that checks maximum string length
//
// Example usage:
// MaxLength(5)("hello") // returns false
// MaxLength(5)("hi") // returns true
func MaxLength(max int) ValidationFunc {
	return func(value any) bool {
		str, ok := value.(string)
		if !ok {
			return false
		}
		return utf8.RuneCountInString(strings.TrimSpace(str)) <= max
	}
}

// Email validates email format
//
// Example usage:
// Email("foo@example.com") // returns true
// Email("invalid-email") // returns false
func Email(value any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	return regexp.MustCompile(rgxEmail).MatchString(str)
}

// Phone validates phone number format
func Phone(value any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	return regexp.MustCompile(rgxPhone).MatchString(str)
}

// Min returns a validation function that checks minimum value
//
// Example usage:
// Min(10)(15) // returns true
// Min(10)(5) // returns false
func Min[T ~int | ~float64](min T) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(T)
		if !ok {
			return false
		}
		return v >= min
	}
}

// Max returns a validation function that checks maximum value
//
// Example usage:
// Max(10)(5) // returns true
// Max(10)(15) // returns false
func Max[T ~int | ~float64](max T) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(T)
		if !ok {
			return false
		}
		return v <= max
	}
}

// Match returns a validation function that checks if a string matches a pattern
//
// Example usage:
// Match(`^[a-zA-Z0-9]+$`)(username) // returns true if username is alphanumeric
// Match(`^[a-zA-Z0-9]+$`)(email) // returns false if email is not alphanumeric
func Match(pattern string) ValidationFunc {
	regex := regexp.MustCompile(pattern)
	return func(value any) bool {
		str, ok := value.(string)
		if !ok {
			return false
		}
		return regex.MatchString(str)
	}
}

// GreaterThan returns a validation function that checks if a value is greater than a specified value
//
// Example usage:
//
// GreaterThan(10)(15) // returns true
// GreaterThan(10)(5) // returns false
func GreaterThan[T Numeric](n T) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(T)
		if !ok {
			return false
		}
		return v > n
	}
}

// LessThan returns a validation function that checks if a value is less than a specified value
//
// Example usage:
// LessThan(10)(5) // returns true
// LessThan(10)(15) // returns false
func LessThan[T Numeric](n T) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(T)
		if !ok {
			return false
		}
		return v < n
	}
}

// GreaterOrEqual returns a validation function that checks if a value is greater than or equal to a specified value
//
// Example usage:
// GreaterOrEqual(10)(15) // returns true
// GreaterOrEqual(10)(10) // returns true
// GreaterOrEqual(10)(5) // returns false
func GreaterOrEqual[T Numeric](n T) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(T)
		if !ok {
			return false
		}
		return v >= n
	}
}

// LessOrEqual returns a validation function that checks if a value is less than or equal to a specified value
//
// Example usage:
// LessOrEqual(10)(5) // returns true
// LessOrEqual(10)(10) // returns true
// LessOrEqual(10)(15) // returns false
func LessOrEqual[T Numeric](n T) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(T)
		if !ok {
			return false
		}
		return v <= n
	}
}
