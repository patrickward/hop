package check

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/exp/constraints"
)

var (
	// RgxEmail is a regular expression for validating email addresses.
	RgxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	// RgxPhone is a regular expression for validating phone numbers.
	RgxPhone = regexp.MustCompile(`^\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})$`)
	// RgxUsername is a regular expression for validating usernames.
	RgxUsername = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9-_]$`)
)

// EqualStrings checks if two strings are equal.
func EqualStrings(value, other string) bool {
	return strings.TrimSpace(value) == strings.TrimSpace(other)
}

// NotBlank checks if a string is not blank.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// NotZero checks if one or more integers are not zero.
func NotZero[T int | int32 | int64](values ...T) bool {
	for i := range values {
		if values[i] == 0 {
			return false
		}
	}
	return true
}

// GreaterThan checks if one integer is greater than another.
func GreaterThan(value, other int) bool {
	return value > other
}

// MinRunes checks if a string has at least n runes.
func MinRunes(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// MaxRunes checks if a string has at most n runes.
func MaxRunes(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// Between checks if a value is between a minimum and maximum value.
func Between[T constraints.Ordered](value, min, max T) bool {
	return value >= min && value <= max
}

// Matches checks if a string matches a regular expression.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// In checks if a value is in a list of values.
func In[T comparable](value T, safelist ...T) bool {
	for i := range safelist {
		if value == safelist[i] {
			return true
		}
	}
	return false
}

// AllIn checks if all values are in a list of values.
func AllIn[T comparable](values []T, safelist ...T) bool {
	for i := range values {
		if !In(values[i], safelist...) {
			return false
		}
	}
	return true
}

// NotIn checks if a value is not in a list of values.
func NotIn[T comparable](value T, blocklist ...T) bool {
	for i := range blocklist {
		if value == blocklist[i] {
			return false
		}
	}
	return true
}

// NoDuplicates checks if a list of values has no duplicates.
func NoDuplicates[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// IsEmail checks if a string is a valid email address.
func IsEmail(value string) bool {
	if len(value) > 254 {
		return false
	}

	return RgxEmail.MatchString(value)
}

// IsPhone checks if a string is a valid phone number.
func IsPhone(value string) bool {
	if len(value) > 15 {
		return false
	}

	return RgxPhone.MatchString(value)
}

// IsUsername checks if a string is a valid username.
func IsUsername(value string) bool {
	if len(value) < 3 || len(value) > 63 {
		return false
	}

	if strings.Contains(value, "__") || strings.Contains(value, "--") {
		return false
	}

	return RgxUsername.MatchString(value)
}

// IsURL checks if a string is a valid URL.
func IsURL(value string) bool {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	return u.Scheme != "" && u.Host != ""
}
