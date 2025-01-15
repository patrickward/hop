package env

import "time"

// TimeLocation gets a time.Location value, returning the default if not set or invalid
func (c *Config) TimeLocation(key string, defaultValue *time.Location) *time.Location {
	if value, found := c.lookup(key); found {
		loc, err := time.LoadLocation(value)
		if err == nil {
			return loc
		}
	}

	return defaultValue
}

// MustTimeLocation gets a time.Location value, panicking if not set or invalid
func (c *Config) MustTimeLocation(key string) *time.Location {
	value := c.get(key)
	if value == "" {
		panic("required environment variable " + c.prefix + key + " not set")
	}
	loc, err := time.LoadLocation(value)
	if err != nil {
		panic("invalid time zone: " + value)
	}
	return loc
}
