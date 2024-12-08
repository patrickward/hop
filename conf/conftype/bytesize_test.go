package conftype_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf/conftype"
)

func TestByteSizeParsesValidString(t *testing.T) {
	var b conftype.ByteSize
	err := b.ParseString("1.5GB")
	require.NoError(t, err)
	assert.Equal(t, conftype.ByteSize(1.5*float64(conftype.GB)), b)
}

func TestByteSizeFailsToParseInvalidString(t *testing.T) {
	var b conftype.ByteSize
	err := b.ParseString("invalid")
	assert.Error(t, err)
}

func TestByteSizeUnmarshalsFromString(t *testing.T) {
	var b conftype.ByteSize
	err := json.Unmarshal([]byte(`"1.5GB"`), &b)
	require.NoError(t, err)
	assert.Equal(t, conftype.ByteSize(1.5*float64(conftype.GB)), b)
}

func TestByteSizeFailsToUnmarshalInvalidJSON(t *testing.T) {
	var b conftype.ByteSize
	err := json.Unmarshal([]byte(`{}`), &b)
	assert.Error(t, err)
}

func TestByteSizeMarshalsToString(t *testing.T) {
	b := conftype.ByteSize(1.5 * float64(conftype.GB))
	data, err := json.Marshal(b)
	require.NoError(t, err)
	assert.JSONEq(t, `"1.50GB"`, string(data))
}

func TestByteSizeStringRepresentation(t *testing.T) {
	b := conftype.ByteSize(1.5 * float64(conftype.GB))
	assert.Equal(t, "1.50GB", b.String())
}
