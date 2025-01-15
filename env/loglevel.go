package env

import "fmt"

type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

// String returns the string representation of the log level and implements the fmt.Stringer interface
func (ll LogLevel) String() string {
	return string(ll)
}

// LogLevel gets a log level value, returning the default if not set or invalid
func (c *Config) LogLevel(key string, defaultValue LogLevel) LogLevel {
	if value, found := c.lookup(key); found {
		switch LogLevel(value) {
		case LogDebug, LogInfo, LogWarn, LogError:
			return LogLevel(value)
		}
	}

	return defaultValue
}

// MustLogLevel gets a log level value, panicking if not set or invalid
func (c *Config) MustLogLevel(key string) LogLevel {
	value := c.get(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s%s not set", c.prefix, key))
	}
	switch LogLevel(value) {
	case LogDebug, LogInfo, LogWarn, LogError:
		return LogLevel(value)
	}
	panic(fmt.Sprintf("invalid log level: %s", value))
}
