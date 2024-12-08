package conftype_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf/conftype"
)

func TestTimeZoneParsesValidString(t *testing.T) {
	var tz conftype.TimeZone
	err := tz.ParseString("America/New_York")
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", tz.String())
}

func TestTimeZoneParsesEmptyString(t *testing.T) {
	var tz conftype.TimeZone
	err := tz.ParseString("")
	require.NoError(t, err)
	assert.Nil(t, tz.Location())
}

func TestTimeZoneFailsToParseInvalidString(t *testing.T) {
	var tz conftype.TimeZone
	err := tz.ParseString("Invalid/TimeZone")
	assert.Error(t, err)
}

func TestTimeZoneUnmarshalsFromStringJSON(t *testing.T) {
	var tz conftype.TimeZone
	err := json.Unmarshal([]byte(`"America/New_York"`), &tz)
	require.NoError(t, err)
	assert.Equal(t, "America/New_York", tz.String())
}

func TestTimeZoneFailsToUnmarshalInvalidJSON(t *testing.T) {
	var tz conftype.TimeZone
	err := json.Unmarshal([]byte(`{}`), &tz)
	assert.Error(t, err)
}

func TestTimeZoneMarshalsToStringJSON(t *testing.T) {
	tz := conftype.TimeZone{}
	_ = tz.ParseString("America/New_York")
	data, err := json.Marshal(tz)
	require.NoError(t, err)
	assert.JSONEq(t, `"America/New_York"`, string(data))
}

func TestTimeZoneMarshalsEmptyToStringJSON(t *testing.T) {
	var tz conftype.TimeZone
	data, err := json.Marshal(tz)
	require.NoError(t, err)
	assert.JSONEq(t, `""`, string(data))
}

func TestTimeZoneStringRepresentation(t *testing.T) {
	tz := conftype.TimeZone{}
	_ = tz.ParseString("America/New_York")
	assert.Equal(t, "America/New_York", tz.String())
}
