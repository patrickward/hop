package url_test

import (
	"html/template"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	hopURL "github.com/patrickward/hop/templates/funcmap/url"
)

func TestUrlSetParam(t *testing.T) {
	tests := []struct {
		key      string
		value    any
		urlStr   string
		expected string
	}{
		{"param", "value", "https://example.com", "https://example.com?param=value"},
		{"param", 123, "https://example.com", "https://example.com?param=123"},
	}

	for _, tt := range tests {
		u, _ := url.Parse(tt.urlStr)
		result := hopURL.FuncMap()["url_set"].(func(string, any, *url.URL) *url.URL)(tt.key, tt.value, u)
		assert.Equal(t, tt.expected, result.String())
	}
}

func TestUrlDelParam(t *testing.T) {
	tests := []struct {
		key      string
		urlStr   string
		expected string
	}{
		{"param", "https://example.com?param=value", "https://example.com"},
		{"param", "https://example.com?param=value&other=123", "https://example.com?other=123"},
	}

	for _, tt := range tests {
		u, _ := url.Parse(tt.urlStr)
		result := hopURL.FuncMap()["url_del"].(func(string, *url.URL) *url.URL)(tt.key, u)
		assert.Equal(t, tt.expected, result.String())
	}
}

func TestUrlToAttr(t *testing.T) {
	tests := []struct {
		urlStr   string
		expected template.HTMLAttr
	}{
		{"https://example.com", template.HTMLAttr("https://example.com")},
		{"https://example.com/path?query=1", template.HTMLAttr("https://example.com/path?query=1")},
	}

	for _, tt := range tests {
		u, _ := url.Parse(tt.urlStr)
		result := hopURL.FuncMap()["url_to_attr"].(func(*url.URL) template.HTMLAttr)(u)
		assert.Equal(t, tt.expected, result)
	}
}

func TestUrlEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com", "https%3A%2F%2Fexample.com"},
		{"https://example.com/path?query=1", "https%3A%2F%2Fexample.com%2Fpath%3Fquery%3D1"},
	}

	for _, tt := range tests {
		result := hopURL.FuncMap()["url_escape"].(func(string) string)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
