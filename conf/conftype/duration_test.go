package conftype_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf/conftype"
)

func TestDurationParsesValidString(t *testing.T) {
	var d conftype.Duration
	err := d.ParseString("1h30m")
	require.NoError(t, err)
	assert.Equal(t, 90*time.Minute, d.Duration)
}

func TestDurationFailsToParseInvalidString(t *testing.T) {
	var d conftype.Duration
	err := d.ParseString("invalid")
	assert.Error(t, err)
}

func TestDurationUnmarshalsFromString(t *testing.T) {
	var d conftype.Duration
	err := json.Unmarshal([]byte(`"1h30m"`), &d)
	require.NoError(t, err)
	assert.Equal(t, 90*time.Minute, d.Duration)
}

func TestDurationUnmarshalsFromNumber(t *testing.T) {
	var d conftype.Duration
	err := json.Unmarshal([]byte(`5400000000000`), &d)
	require.NoError(t, err)
	assert.Equal(t, 90*time.Minute, d.Duration)
}

func TestDurationFailsToUnmarshalInvalidJSON(t *testing.T) {
	var d conftype.Duration
	err := json.Unmarshal([]byte(`{}`), &d)
	assert.Error(t, err)
}

func TestDurationMarshalsToString(t *testing.T) {
	d := conftype.Duration{Duration: 90 * time.Minute}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	assert.JSONEq(t, `"1h30m0s"`, string(b))
}

func TestDurationStringRepresentation(t *testing.T) {
	d := conftype.Duration{Duration: 90 * time.Minute}
	assert.Equal(t, "1h30m0s", d.String())
}
