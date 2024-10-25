package hopconfig

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a wrapper around time.Duration that supports JSON marshaling/unmarshaling
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements the json.Unmarshaler interface
//
//goland:noinspection ALL
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid duration: %v", v)
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface
//
//goland:noinspection ALL
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}
