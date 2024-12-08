package conftype

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type ByteSize int64

const (
	_           = iota
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
)

// ParseString handles environment variables and default values
func (b *ByteSize) ParseString(s string) error {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	var multiplier ByteSize = 1

	// Handle unit suffixes
	switch {
	case strings.HasSuffix(s, "PB"):
		multiplier = PB
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "TB"):
		multiplier = TB
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "GB"):
		multiplier = GB
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "MB"):
		multiplier = MB
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "KB"):
		multiplier = KB
		s = s[:len(s)-2]
	case strings.HasSuffix(s, "B"):
		s = s[:len(s)-1]
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid byte size %q: %w", s, err)
	}

	*b = ByteSize(value * float64(multiplier))
	return nil
}

// UnmarshalJSON implements json.Unmarshaler
func (b *ByteSize) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return b.ParseString(s)
}

// MarshalJSON implements json.Marshaler
func (b ByteSize) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

// String implements fmt.Stringer
func (b ByteSize) String() string {
	switch {
	case b >= PB:
		return fmt.Sprintf("%.2fPB", float64(b)/float64(PB))
	case b >= TB:
		return fmt.Sprintf("%.2fTB", float64(b)/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.2fGB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2fMB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2fKB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
