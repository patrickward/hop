package conftype_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf/conftype"
)

func TestLogLevelParsesValidString(t *testing.T) {
	var l conftype.LogLevel
	err := l.ParseString("debug")
	require.NoError(t, err)
	assert.Equal(t, conftype.LogDebug, l)
}

func TestLogLevelFailsToParseInvalidString(t *testing.T) {
	var l conftype.LogLevel
	err := l.ParseString("invalid")
	assert.Error(t, err)
}

func TestLogLevelUnmarshalsFromString(t *testing.T) {
	var l conftype.LogLevel
	err := json.Unmarshal([]byte(`"info"`), &l)
	require.NoError(t, err)
	assert.Equal(t, conftype.LogInfo, l)
}

func TestLogLevelFailsToUnmarshalInvalidJSON(t *testing.T) {
	var l conftype.LogLevel
	err := json.Unmarshal([]byte(`{}`), &l)
	assert.Error(t, err)
}

func TestLogLevelMarshalsToString(t *testing.T) {
	l := conftype.LogLevel("warn")
	data, err := json.Marshal(l)
	require.NoError(t, err)
	assert.JSONEq(t, `"warn"`, string(data))
}

func TestLogLevelStringRepresentation(t *testing.T) {
	l := conftype.LogLevel("error")
	assert.Equal(t, "error", l.String())
}
