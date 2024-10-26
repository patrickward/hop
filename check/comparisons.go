package check

import (
	"strings"
	"time"

	"golang.org/x/exp/constraints"
)

// Before checks if a time is before another
func Before(t time.Time) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(time.Time)
		if !ok {
			return false
		}
		return v.Before(t)
	}
}

// After checks if a time is after another
func After(t time.Time) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(time.Time)
		if !ok {
			return false
		}
		return v.After(t)
	}
}

// BetweenTime checks if a value is between two other values
//
// Example usage:
// BetweenTime(time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))(time.Now()) // returns true
// BetweenTime(time.Now().Add(-1*time.Hour), time.Now().Add(-30*time.Minute))(time.Now()) // returns false
func BetweenTime(start, end time.Time) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(time.Time)
		if !ok {
			return false
		}
		return v.After(start) && v.Before(end)
	}
}

// Between checks if a value is between a minimum and maximum value
// Example usage:
//
// Between(10, 20)(15) // returns true
// Between(10, 20)(25) // returns false
func Between[T constraints.Ordered](min, max T) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(T)
		if !ok {
			return false
		}
		return v >= min && v <= max
	}
}

// Equal checks if values are equal
// Example usage:
// Equal(10)(10) // returns true
// Equal(10)(5) // returns false
func Equal(other any) ValidationFunc {
	return func(value any) bool {
		return value == other
	}
}

// EqualStrings checks if two strings are equal
//
// Example usage:
// EqualStrings("test", "test") // returns true
// EqualStrings("test", "TEST") // returns false
func EqualStrings(value, other string) bool {
	return strings.TrimSpace(value) == strings.TrimSpace(other)
}

// In checks if a value is in a set of allowed values
//
// Example usage:
// In(1, 2, 3)(2) // returns true
// In(1, 2, 3)(4) // returns false
// In("a", "b", "c")("b") // returns true
// In("a", "b", "c")("d") // returns false
func In[T comparable](allowed ...T) ValidationFunc {
	return func(value any) bool {
		v, ok := value.(T)
		if !ok {
			return false
		}
		for _, a := range allowed {
			if v == a {
				return true
			}
		}
		return false
	}
}

// AllIn checks if all values in a slice are in a set of allowed values
//
// Example usage:
// AllIn(1, 2, 3)([]int{1, 2}) // returns true
// AllIn(1, 2, 3)([]int{1, 4}) // returns false
// AllIn("a", "b", "c")([]string{"a", "b"}) // returns true
// AllIn("a", "b", "c")([]string{"a", "d"}) // returns false
func AllIn[T comparable](allowed ...T) ValidationFunc {
	return func(value any) bool {
		values, ok := value.([]T)
		if !ok {
			return false
		}

		// Create a map for O(1) lookups
		allowedMap := make(map[T]struct{}, len(allowed))
		for _, a := range allowed {
			allowedMap[a] = struct{}{}
		}

		for _, v := range values {
			if _, exists := allowedMap[v]; !exists {
				return false
			}
		}
		return true
	}
}

// NoDuplicates checks if a slice contains any duplicate values
//
// Example usage:
// NoDuplicates()([]int{1, 2, 3}) // returns true
// NoDuplicates()([]int{1, 2, 2}) // returns false
func NoDuplicates[T comparable]() ValidationFunc {
	return func(value any) bool {
		values, ok := value.([]T)
		if !ok {
			return false
		}

		seen := make(map[T]struct{}, len(values))
		for _, v := range values {
			if _, exists := seen[v]; exists {
				return false
			}
			seen[v] = struct{}{}
		}
		return true
	}
}
