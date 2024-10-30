package decode

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// Query decodes the query string values in an HTTP request into a struct.
func Query(r *http.Request, dst any) error {
	return decodeURLValues(r.URL.Query(), dst)
}

// QueryString returns the value of a query string parameter in an HTTP request.
func QueryString(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}

// QueryInt64 returns the value of a query string parameter as an int64 in an HTTP request.
func QueryInt64(r *http.Request, key string) int64 {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0
	}
	parseInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return parseInt
}

// QueryFloat64 returns the value of a query string parameter as a float64 in an HTTP request.
func QueryFloat64(r *http.Request, key string) float64 {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0
	}
	parseFloat, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}

	return parseFloat
}

// QueryBool returns the value of a query string parameter as a bool in an HTTP request.
func QueryBool(r *http.Request, key string) bool {
	value := r.URL.Query().Get(key)
	if value == "" {
		return false
	}
	parseBool, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return parseBool
}

// QueryIsTrue checks if a query string parameter is set to a truthy value.
func QueryIsTrue(r *http.Request, key string) bool {
	allowedValues := []string{"true", "t", "1", "yes", "on"}
	value := r.URL.Query().Get(key)
	return slices.Contains(allowedValues, value)
}
