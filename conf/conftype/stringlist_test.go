package conftype_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf/conftype"
)

func TestStringListParsesValidString(t *testing.T) {
	var l conftype.StringList
	err := l.ParseString("a,b,c")
	require.NoError(t, err)
	assert.Equal(t, conftype.StringList{"a", "b", "c"}, l)
}

func TestStringListParsesEmptyString(t *testing.T) {
	var l conftype.StringList
	err := l.ParseString("")
	require.NoError(t, err)
	assert.Nil(t, l)
}

func TestStringListParsesStringWithSpaces(t *testing.T) {
	var l conftype.StringList
	err := l.ParseString(" a , b , c ")
	require.NoError(t, err)
	assert.Equal(t, conftype.StringList{"a", "b", "c"}, l)
}

func TestStringListUnmarshalsFromArrayJSON(t *testing.T) {
	var l conftype.StringList
	err := json.Unmarshal([]byte(`["a", "b", "c"]`), &l)
	require.NoError(t, err)
	assert.Equal(t, conftype.StringList{"a", "b", "c"}, l)
}

func TestStringListUnmarshalsFromStringJSON(t *testing.T) {
	var l conftype.StringList
	err := json.Unmarshal([]byte(`"a,b,c"`), &l)
	require.NoError(t, err)
	assert.Equal(t, conftype.StringList{"a", "b", "c"}, l)
}

func TestStringListFailsToUnmarshalInvalidJSON(t *testing.T) {
	var l conftype.StringList
	err := json.Unmarshal([]byte(`{}`), &l)
	assert.Error(t, err)
}

func TestStringListMarshalsToArrayJSON(t *testing.T) {
	l := conftype.StringList{"a", "b", "c"}
	data, err := json.Marshal(l)
	require.NoError(t, err)
	assert.JSONEq(t, `["a", "b", "c"]`, string(data))
}

func TestStringListMarshalsEmptyListToArrayJSON(t *testing.T) {
	var l conftype.StringList
	data, err := json.Marshal(l)
	require.NoError(t, err)
	assert.JSONEq(t, `[]`, string(data))
}

func TestStringListStringRepresentation(t *testing.T) {
	l := conftype.StringList{"a", "b", "c"}
	assert.Equal(t, "a,b,c", l.String())
}
