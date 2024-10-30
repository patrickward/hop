package decode

import (
	"net/http"
	"strconv"
	"strings"
)

func URLParamInt64(r *http.Request, key string) int64 {
	value := r.PathValue(key)
	if value == "" {
		return 0
	}
	parseInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return parseInt
}

func URLParamString(r *http.Request, key string) string {
	return strings.TrimSpace(r.PathValue(key))
}

func URLParamFloat64(r *http.Request, key string) float64 {
	value := r.PathValue(key)
	if value == "" {
		return 0
	}
	parseFloat, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}

	return parseFloat
}

func URLParamBool(r *http.Request, key string) bool {
	value := r.PathValue(key)
	if value == "" {
		return false
	}
	parseBool, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return parseBool
}
