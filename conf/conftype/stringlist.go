package conftype

import (
	"encoding/json"
	"strings"
)

// StringList represents a list of strings, typically comma-separated in config
type StringList []string

// ParseString implements StringParser
//
//goland:noinspection GoMixedReceiverTypes
func (l *StringList) ParseString(s string) error {
	if s == "" {
		*l = nil
		return nil
	}

	// Split and trim each value
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			result = append(result, v)
		}
	}

	*l = result
	return nil
}

// UnmarshalJSON implements json.Unmarshaler
//
//goland:noinspection GoMixedReceiverTypes
func (l *StringList) UnmarshalJSON(data []byte) error {
	// Try array format first
	var values []string
	if err := json.Unmarshal(data, &values); err == nil {
		*l = values
		return nil
	}

	// Fall back to string format
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return l.ParseString(s)
}

// MarshalJSON implements json.Marshaler
//
//goland:noinspection GoMixedReceiverTypes
func (l StringList) MarshalJSON() ([]byte, error) {
	if l == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(l))
}

// String implements fmt.Stringer
//
//goland:noinspection GoMixedReceiverTypes
func (l StringList) String() string {
	return strings.Join(l, ",")
}
