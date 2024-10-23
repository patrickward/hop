package httpgo

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-playground/form/v4"
)

var decoder = form.NewDecoder()

// DecodeForm decodes the form values in an HTTP request into a struct.
func DecodeForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	return decodeURLValues(r.Form, dst)
}

// DecodePostForm decodes the POST form values in an HTTP request into a struct.
func DecodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	return decodeURLValues(r.PostForm, dst)
}

// DecodeQueryString decodes the query string values in an HTTP request into a struct.
func DecodeQueryString(r *http.Request, dst any) error {
	return decodeURLValues(r.URL.Query(), dst)
}

func decodeURLValues(v url.Values, dst any) error {
	err := decoder.Decode(dst, v)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
	}

	return err
}
