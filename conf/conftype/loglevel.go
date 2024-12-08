package conftype

import (
	"encoding/json"
	"fmt"
	"strings"
)

type LogLevel string

const (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

// ParseString handles environment variables and default values
//
//goland:noinspection GoMixedReceiverTypes
func (l *LogLevel) ParseString(s string) error {
	switch strings.ToLower(s) {
	case "debug", "info", "warn", "error":
		*l = LogLevel(strings.ToLower(s))
		return nil
	default:
		return fmt.Errorf("invalid log level %q: must be one of: debug, info, warn, error", s)
	}
}

// UnmarshalJSON implements json.Unmarshaler
//
//goland:noinspection GoMixedReceiverTypes
func (l *LogLevel) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return l.ParseString(s)
}

// MarshalJSON implements json.Marshaler
//
//goland:noinspection GoMixedReceiverTypes
func (l LogLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(l))
}

// String implements fmt.Stringer
//
//goland:noinspection GoMixedReceiverTypes
func (l LogLevel) String() string {
	return string(l)
}
