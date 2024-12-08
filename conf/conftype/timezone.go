package conftype

import (
	"encoding/json"
	"fmt"
	"time"
)

// TimeZone represents a validated time zone name
type TimeZone struct {
	location *time.Location
}

// ParseString handles environment variables and default values
//
//goland:noinspection GoMixedReceiverTypes
func (tz *TimeZone) ParseString(s string) error {
	if s == "" {
		tz.location = nil
		return nil
	}

	loc, err := time.LoadLocation(s)
	if err != nil {
		return fmt.Errorf("invalid time zone %q: %w", s, err)
	}

	tz.location = loc
	return nil
}

// UnmarshalJSON implements json.Unmarshaler
//
//goland:noinspection GoMixedReceiverTypes
func (tz *TimeZone) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return tz.ParseString(s)
}

// MarshalJSON implements json.Marshaler
//
//goland:noinspection GoMixedReceiverTypes
func (tz TimeZone) MarshalJSON() ([]byte, error) {
	return json.Marshal(tz.String())
}

// String implements fmt.Stringer
//
//goland:noinspection GoMixedReceiverTypes
func (tz TimeZone) String() string {
	if tz.location == nil {
		return ""
	}
	return tz.location.String()
}

// Location returns the time.Location
//
//goland:noinspection GoMixedReceiverTypes
func (tz TimeZone) Location() *time.Location {
	return tz.location
}
