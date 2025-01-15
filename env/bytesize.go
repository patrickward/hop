package env

import (
	"fmt"
	"strconv"
	"strings"
)

func (c *Config) GetByteSize(key string, defaultValue int) int {
	if value, found := c.lookup(key); found {
		result, err := parseByteSize(value)
		if err == nil {
			return result
		}
	}

	return defaultValue
}

func (c *Config) MustByteSize(key string) int {
	if value, found := c.lookup(key); found {
		result, err := parseByteSize(value)
		if err == nil {
			return result
		}
	}

	panic(fmt.Sprintf("required environment variable %s%s not set", c.prefix, key))
}

func parseByteSize(s string) (int, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	var multiplier int = 1

	// Handle unit suffixes
	switch {
	case strings.HasSuffix(s, "PB"):
		multiplier = 1 << (10 * 5)
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "TB"):
		multiplier = 1 << (10 * 4)
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "GB"):
		multiplier = 1 << (10 * 3)
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "MB"):
		multiplier = 1 << (10 * 2)
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "KB"):
		multiplier = 1 << 10
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "B"):
		s = s[:len(s)-1]
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size %q: %w", s, err)
	}

	return int(value * float64(multiplier)), nil
}
