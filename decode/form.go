package decode

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-playground/form/v4"
)

var formDecoder = form.NewDecoder()

// Form decodes the form values in an HTTP request into a struct.
func Form(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	return decodeURLValues(r.Form, dst)
}

// PostForm decodes the POST form values in an HTTP request into a struct.
func PostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	return decodeURLValues(r.PostForm, dst)
}

func decodeURLValues(v url.Values, dst any) error {
	err := formDecoder.Decode(dst, v)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
	}

	return err
}
