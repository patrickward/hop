package env

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Getter is a function type for custom environment variable parsing
type Getter func(value string) (any, error)

// Config provides access to environment variables with a consistent prefix
type Config struct {
	prefix  string
	getters map[string]Getter
	mu      sync.RWMutex
}

// New creates a new Config instance with the given prefix
func New(prefix string) *Config {
	return &Config{
		prefix:  prefix,
		getters: make(map[string]Getter),
	}
}

// SetPrefix sets the prefix for the Config instance
func (c *Config) SetPrefix(prefix string) {
	c.prefix = prefix
}

// Set sets the environment variable value for the given key
func (c *Config) Set(key, value string) {
	_ = os.Setenv(c.prefix+key, value)
}

// RegisterGetter registers a new getter function for a custom type
func (c *Config) RegisterGetter(name string, getter Getter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.getters[name] = getter
}

// Get retrieves a value using a registered getter
func (c *Config) Get(key, getterName string, defaultValue any) any {
	c.mu.RLock()
	getter, ok := c.getters[getterName]
	c.mu.RUnlock()

	if !ok {
		panic(fmt.Sprintf("getter %q not registered", getterName))
	}

	if value := c.get(key); value != "" {
		if result, err := getter(value); err == nil {
			return result
		}
	}
	return defaultValue
}

// MustGet retrieves a value using a registered getter, panicking if not set or invalid
func (c *Config) MustGet(key, getterName string) any {
	c.mu.RLock()
	getter, ok := c.getters[getterName]
	c.mu.RUnlock()

	if !ok {
		panic(fmt.Sprintf("getter %q not registered", getterName))
	}

	value := c.get(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s%s not set", c.prefix, key))
	}

	result, err := getter(value)
	if err != nil {
		panic(fmt.Sprintf("invalid value for %s%s: %v", c.prefix, key, err))
	}
	return result
}

// get returns the environment variable value for the given key, applying the prefix
func (c *Config) get(key string) string {
	return os.Getenv(c.prefix + key)
}

// String gets a string value, returning the default if not set
func (c *Config) String(key string, defaultValue string) string {
	if value := c.get(key); value != "" {
		return value
	}
	return defaultValue
}

// MustString gets a string value, panicking if not set
func (c *Config) MustString(key string) string {
	if value := c.get(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("required environment variable %s%s not set", c.prefix, key))
}

// Int gets an integer value, returning the default if not set or invalid
func (c *Config) Int(key string, defaultValue int) int {
	if value := c.get(key); value != "" {
		if i, err := parseInt(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// MustInt gets an integer value, panicking if not set or invalid
func (c *Config) MustInt(key string) int {
	value := c.get(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s%s not set", c.prefix, key))
	}
	i, err := parseInt(value)
	if err != nil {
		panic(fmt.Sprintf("invalid integer value for %s%s: %s", c.prefix, key, value))
	}
	return i
}

// Bool gets a boolean value, returning the default if not set
func (c *Config) Bool(key string, defaultValue bool) bool {
	if value := c.get(key); value != "" {
		return parseBool(value)
	}
	return defaultValue
}

// MustBool gets a boolean value, panicking if not set or invalid
func (c *Config) MustBool(key string) bool {
	value := c.get(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s%s not set", c.prefix, key))
	}
	return parseBool(value)
}

// Duration gets a duration value, returning the default if not set or invalid
func (c *Config) Duration(key string, defaultValue time.Duration) time.Duration {
	if value := c.get(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

// MustDuration gets a duration value, panicking if not set or invalid
func (c *Config) MustDuration(key string) time.Duration {
	value := c.get(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s%s not set", c.prefix, key))
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Sprintf("invalid duration value for %s%s: %s", c.prefix, key, value))
	}
	return d
}

// StringSlice gets a slice of strings, splitting on comma, returning the default if not set
func (c *Config) StringSlice(key string, defaultValue []string) []string {
	if value := c.get(key); value != "" {
		return splitAndTrim(value)
	}
	return defaultValue
}

// MustStringSlice gets a slice of strings, panicking if not set
func (c *Config) MustStringSlice(key string) []string {
	value := c.get(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s%s not set", c.prefix, key))
	}
	return splitAndTrim(value)
}

// Helpers

func parseInt(value string) (int, error) {
	var i int
	_, err := fmt.Sscanf(value, "%d", &i)
	return i, err
}

func parseBool(value string) bool {
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

func splitAndTrim(value string) []string {
	var values []string
	for _, v := range strings.Split(value, ",") {
		if v = strings.TrimSpace(v); v != "" {
			values = append(values, v)
		}
	}
	return values
}
