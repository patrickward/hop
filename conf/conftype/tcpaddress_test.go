package conftype_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf/conftype"
)

func TestTCPAddressParsesValidString(t *testing.T) {
	var a conftype.TCPAddress
	err := a.ParseString("localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "localhost", a.Host)
	assert.Equal(t, 8080, a.Port)
}

func TestTCPAddressParsesEmptyString(t *testing.T) {
	var a conftype.TCPAddress
	err := a.ParseString("")
	require.NoError(t, err)
	assert.Equal(t, "", a.Host)
	assert.Equal(t, 0, a.Port)
}

func TestTCPAddressFailsToParseInvalidString(t *testing.T) {
	var a conftype.TCPAddress
	err := a.ParseString("invalid")
	assert.Error(t, err)
}

func TestTCPAddressFailsToParseInvalidPort(t *testing.T) {
	var a conftype.TCPAddress
	err := a.ParseString("localhost:invalid")
	assert.Error(t, err)
}

func TestTCPAddressFailsToParseOutOfRangePort(t *testing.T) {
	var a conftype.TCPAddress
	err := a.ParseString("localhost:70000")
	assert.Error(t, err)
}

func TestTCPAddressUnmarshalsFromString(t *testing.T) {
	var a conftype.TCPAddress
	err := json.Unmarshal([]byte(`"localhost:8080"`), &a)
	require.NoError(t, err)
	assert.Equal(t, "localhost", a.Host)
	assert.Equal(t, 8080, a.Port)
}

func TestTCPAddressFailsToUnmarshalInvalidJSON(t *testing.T) {
	var a conftype.TCPAddress
	err := json.Unmarshal([]byte(`{}`), &a)
	assert.Error(t, err)
}

func TestTCPAddressMarshalsToString(t *testing.T) {
	a := conftype.TCPAddress{Host: "localhost", Port: 8080}
	data, err := json.Marshal(a)
	require.NoError(t, err)
	assert.JSONEq(t, `"localhost:8080"`, string(data))
}

func TestTCPAddressStringRepresentation(t *testing.T) {
	a := conftype.TCPAddress{Host: "localhost", Port: 8080}
	assert.Equal(t, "localhost:8080", a.String())
}

func TestTCPAddressNetworkAddress(t *testing.T) {
	a := conftype.TCPAddress{Host: "localhost", Port: 8080}
	assert.Equal(t, "localhost:8080", a.NetworkAddress())
}
