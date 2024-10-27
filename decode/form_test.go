package decode_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/decode"
)

type TestData struct {
	Field string `form:"field"`
}

func TestForm(t *testing.T) {
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
			err := decode.Form(req, &dst)

			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, dst)
		})
	}
}

func TestPostForm(t *testing.T) {
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
			err := decode.PostForm(req, &dst)

			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, dst)
		})
	}
}
