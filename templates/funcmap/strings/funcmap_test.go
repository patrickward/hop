package strings_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/strings"
)

func TestTitleize(t *testing.T) {
	assert.Equal(t, "Hello, World", strings.Titleize("hello, world"))
	assert.Equal(t, "Hello, World", strings.Titleize("Hello, World"))
	assert.Equal(t, "Hello, World", strings.Titleize("HELLO, WORLD"))
	assert.Equal(t, "Hello, World", strings.Titleize("hELLO, wORLD"))
}
