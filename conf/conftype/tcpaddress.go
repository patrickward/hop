package conftype

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
)

type TCPAddress struct {
	Host string
	Port int
}

// ParseString handles environment variables and default values
//
//goland:noinspection GoMixedReceiverTypes
func (a *TCPAddress) ParseString(s string) error {
	if s == "" {
		return nil
	}

	// Split host and port
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return fmt.Errorf("invalid TCP address %q: %w", s, err)
	}

	// Parse port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port in TCP address %q: %w", s, err)
	}

	// Validate port range
	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d out of range [1-65535]", port)
	}

	// If host is empty, use localhost
	if host == "" {
		host = "localhost"
	}

	a.Host = host
	a.Port = port
	return nil
}

// UnmarshalJSON implements json.Unmarshaler
//
//goland:noinspection GoMixedReceiverTypes
func (a *TCPAddress) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return a.ParseString(s)
}

// MarshalJSON implements json.Marshaler
//
//goland:noinspection GoMixedReceiverTypes
func (a TCPAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// String implements fmt.Stringer
//
//goland:noinspection GoMixedReceiverTypes
func (a TCPAddress) String() string {
	if a.Host == "" && a.Port == 0 {
		return ""
	}
	return net.JoinHostPort(a.Host, strconv.Itoa(a.Port))
}

// NetworkAddress returns the address in host:port format
//
//goland:noinspection GoMixedReceiverTypes
func (a TCPAddress) NetworkAddress() string {
	return a.String()
}
