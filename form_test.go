package httpgo_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/go-playground/form/v4"
	"github.com/stretchr/testify/assert"

	"github.com/patrickward/httpgo"
)

type TestData struct {
	Field string `form:"field"`
}

var decoder = form.NewDecoder()

func TestDecodeForm(t *testing.T) {
	tt := []struct {
		name     string
		url      string
		expected TestData
		err      error
	}{
		{"Valid Form", "http://localhost?field=value", TestData{Field: "value"}, nil},
		{"Invalid Form", "http://localhost?invalid_field=value", TestData{}, nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("field", tc.expected.Field)

			req, _ := http.NewRequest("GET", tc.url, nil)
			req.Form = form
			var dst TestData
			err := httpgo.DecodeForm(req, &dst)

			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, dst)
		})
	}
}

func TestDecodePostForm(t *testing.T) {
	tt := []struct {
		name     string
		url      string
		expected TestData
		err      error
	}{
		{"Valid Post Form", "http://localhost?field=value", TestData{Field: "value"}, nil},
		{"Invalid Post Form", "http://localhost?invalid_field=value", TestData{}, nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("field", tc.expected.Field)

			req, _ := http.NewRequest("POST", tc.url, nil)
			req.PostForm = form
			var dst TestData
			err := httpgo.DecodePostForm(req, &dst)

			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, dst)
		})
	}
}

func TestDecodeQueryString(t *testing.T) {
	tt := []struct {
		name     string
		url      string
		expected TestData
		err      error
	}{
		{"Valid Query String", "http://localhost?field=value", TestData{Field: "value"}, nil},
		{"Invalid Query String", "http://localhost?invalid_field=value", TestData{}, nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.url, nil)
			var dst TestData
			err := httpgo.DecodeQueryString(req, &dst)

			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, dst)
		})
	}
}
