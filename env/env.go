package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// GetString returns the value of the environment variable key or defaultValue if it does not exist
func GetString(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	return value
}

// GetInt returns the value of the environment variable key as an int or defaultValue if it does not exist
func GetInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}

	return intValue
}

// GetBool returns the value of the environment variable key as a bool or defaultValue if it does not exist
func GetBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		panic(err)
	}

	return boolValue
}

// GetStringSlice returns the value of the environment variable key as a slice of strings or defaultValue if it does not exist
func GetStringSlice(key string, defaultValue []string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	var cleanedValues []string
	for _, v := range strings.Split(value, ",") {
		if v != "" {
			cleanedValues = append(cleanedValues, strings.TrimSpace(v))
		}
	}

	return cleanedValues
}

// GetDuration returns the value of the environment variable key as a duration or defaultValue if it does not exist
func GetDuration(key string, duration time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists {
		return duration
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		panic(err)
	}

	return d
}
