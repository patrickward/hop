package attr_test

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/attr"
)

func classBool(b bool, value string) any {
	if b {
		return value
	}

	return nil
}

func TestAttrClass(t *testing.T) {
	tests := []struct {
		pairs    any
		expected template.HTMLAttr
	}{
		{[]any{"test"}, template.HTMLAttr("class=\"test\"")},
		{[]any{"test", "test2"}, template.HTMLAttr("class=\"test test2\"")},
		{[]any{"test", "test2", "test3  "}, template.HTMLAttr("class=\"test test2 test3\"")},
		{[]any{" test ", classBool(true, "test2")}, template.HTMLAttr("class=\"test test2\"")},
		{[]any{" test ", classBool(false, "test2"), classBool(true, "test3")}, template.HTMLAttr("class=\"test test3\"")},
		{[]any{" ", classBool(false, "test2"), classBool(true, "test3")}, template.HTMLAttr("class=\"test3\"")},
		{[]any{" "}, template.HTMLAttr("")},
	}

	for _, tt := range tests {
		result := attr.FuncMap()["attr_class"].(func(...any) template.HTMLAttr)(tt.pairs.([]any)...)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSafeAttr(t *testing.T) {
	tests := []struct {
		input    string
		expected template.HTMLAttr
	}{
		{"class=\"test\"", template.HTMLAttr("class=\"test\"")},
		{"onclick=\"alert('xss')\"", template.HTMLAttr("onclick=\"alert('xss')\"")},
	}

	for _, tt := range tests {
		result := attr.FuncMap()["attr_safe"].(func(string) template.HTMLAttr)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSelectedAttr(t *testing.T) {
	tests := []struct {
		value    string
		current  any
		expected template.HTMLAttr
	}{
		{"option1", "option1", template.HTMLAttr("selected")},
		{"option1", "option2", template.HTMLAttr("")},
		{"option1", []string{"option1", "option2"}, template.HTMLAttr("selected")},
		{"option1", []string{"option2", "option3"}, template.HTMLAttr("")},
	}

	for _, tt := range tests {
		result := attr.FuncMap()["attr_selected"].(func(any, string) template.HTMLAttr)(tt.current, tt.value)
		assert.Equal(t, tt.expected, result)
	}
}

func TestCheckedAttr(t *testing.T) {
	tests := []struct {
		value    string
		current  any
		expected template.HTMLAttr
	}{
		{"option1", "option1", template.HTMLAttr("checked")},
		{"option1", "option2", template.HTMLAttr("")},
		{"option1", []string{"option1", "option2"}, template.HTMLAttr("checked")},
		{"option1", []string{"option2", "option3"}, template.HTMLAttr("")},
	}

	for _, tt := range tests {
		result := attr.FuncMap()["attr_checked"].(func(any, string) template.HTMLAttr)(tt.current, tt.value)
		assert.Equal(t, tt.expected, result)
	}
}

func TestDisabledAttr(t *testing.T) {
	tests := []struct {
		value    string
		current  any
		expected template.HTMLAttr
	}{
		{"option1", "option1", template.HTMLAttr("disabled")},
		{"option1", "option2", template.HTMLAttr("")},
		{"option1", []string{"option1", "option2"}, template.HTMLAttr("disabled")},
		{"option1", []string{"option2", "option3"}, template.HTMLAttr("")},
	}

	for _, tt := range tests {
		result := attr.FuncMap()["attr_disabled"].(func(any, string) template.HTMLAttr)(tt.current, tt.value)
		assert.Equal(t, tt.expected, result)
	}
}

func TestReadonlyAttr(t *testing.T) {
	tests := []struct {
		value    string
		current  any
		expected template.HTMLAttr
	}{
		{"option1", "option1", template.HTMLAttr("readonly")},
		{"option1", "option2", template.HTMLAttr("")},
		{"option1", []string{"option1", "option2"}, template.HTMLAttr("readonly")},
		{"option1", []string{"option2", "option3"}, template.HTMLAttr("")},
	}

	for _, tt := range tests {
		result := attr.FuncMap()["attr_readonly"].(func(any, string) template.HTMLAttr)(tt.current, tt.value)
		assert.Equal(t, tt.expected, result)
	}
}
