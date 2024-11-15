package env

import (
	"os"
	"strconv"
	"strings"
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
func GetStringSlice(key, defaultValue string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return []string{defaultValue}
	}

	var cleanedValues []string
	for _, v := range strings.Split(value, ",") {
		if v != "" {
			cleanedValues = append(cleanedValues, strings.TrimSpace(v))
		}
	}

	return cleanedValues
}
