package html_test

import (
	"html/template"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/templates/funcmap/html"
)

func TestSafeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected template.HTML
	}{
		{"<div>Test</div>", template.HTML("<div>Test</div>")},
		{"<script>alert('xss')</script>", template.HTML("<script>alert('xss')</script>")},
	}

	for _, tt := range tests {
		result := html.FuncMap()["htm_safe"].(func(string) template.HTML)(tt.input)
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
		result := html.FuncMap()["htm_attr"].(func(string) template.HTMLAttr)(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSelectedAttr(t *testing.T) {
	tests := []struct {
		value    string
		current  string
		expected template.HTMLAttr
	}{
		{"option1", "option1", template.HTMLAttr("selected")},
		{"option1", "option2", template.HTMLAttr("")},
	}

	for _, tt := range tests {
		result := html.FuncMap()["htm_selected"].(func(string, string) template.HTMLAttr)(tt.value, tt.current)
		assert.Equal(t, tt.expected, result)
	}
}

func TestUrlSetParam(t *testing.T) {
	tests := []struct {
		key      string
		value    any
		urlStr   string
		expected string
	}{
		{"param", "value", "http://example.com", "http://example.com?param=value"},
		{"param", 123, "http://example.com", "http://example.com?param=123"},
	}

	for _, tt := range tests {
		u, _ := url.Parse(tt.urlStr)
		result := html.FuncMap()["htm_urlSetParam"].(func(string, any, *url.URL) *url.URL)(tt.key, tt.value, u)
		assert.Equal(t, tt.expected, result.String())
	}
}

func TestUrlDelParam(t *testing.T) {
	tests := []struct {
		key      string
		urlStr   string
		expected string
	}{
		{"param", "http://example.com?param=value", "http://example.com"},
		{"param", "http://example.com?param=value&other=123", "http://example.com?other=123"},
	}

	for _, tt := range tests {
		u, _ := url.Parse(tt.urlStr)
		result := html.FuncMap()["htm_urlDelParam"].(func(string, *url.URL) *url.URL)(tt.key, u)
		assert.Equal(t, tt.expected, result.String())
	}
}

func TestUrlToAttr(t *testing.T) {
	tests := []struct {
		urlStr   string
		expected template.HTMLAttr
	}{
		{"http://example.com", template.HTMLAttr("http://example.com")},
		{"https://example.com/path?query=1", template.HTMLAttr("https://example.com/path?query=1")},
	}

	for _, tt := range tests {
		u, _ := url.Parse(tt.urlStr)
		result := html.FuncMap()["htm_urlToAttr"].(func(*url.URL) template.HTMLAttr)(u)
		assert.Equal(t, tt.expected, result)
	}
}
