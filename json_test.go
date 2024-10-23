package httpgo_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/httpgo"
)

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name    string
		srcJSON string
		target  interface{}
		want    interface{}
		err     error
	}{
		{
			name:    "EmptyJSON",
			srcJSON: `{}`,
			target:  &struct{}{},
			want:    &struct{}{},
		},
		{
			name: "ValidJSON",
			srcJSON: `{
				"number": 123,
				"string": "abc",
				"boolean": true,
				"array": [1, 2, 3],
				"object": {"key": "value"}
			}`,
			target: &struct {
				Number  int
				String  string
				Boolean bool
				Array   []int
				Object  map[string]string
			}{},
			want: &struct {
				Number  int
				String  string
				Boolean bool
				Array   []int
				Object  map[string]string
			}{
				Number:  123,
				String:  "abc",
				Boolean: true,
				Array:   []int{1, 2, 3},
				Object:  map[string]string{"key": "value"},
			},
		},
		{
			name:    "IncorrectType",
			srcJSON: `{"number": "abc"}`,
			target:  &map[string]int{},
			err:     errors.New("body contains incorrect JSON type"),
		},
		{
			name:    "UnknownField is OK in DecodeJSON (non-strict)",
			srcJSON: `{"unknown": 123}`,
			target:  &struct{ Known string }{},
			err:     nil,
		},
		{
			name:    "DisallowedFieldInstinct",
			srcJSON: `{"data": 123}`,
			target:  &struct{ Data interface{} }{},
			want:    &struct{ Data interface{} }{Data: float64(123)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.srcJSON))
			w := httptest.NewRecorder()
			err := httpgo.DecodeJSON(w, r, test.target)
			if err != nil && test.err == nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && test.err != nil {
				t.Errorf("Expected error: %v", test.err)
			}
			if test.want != nil {
				assert.EqualValues(t, test.target, test.want, "Expected value: %v, but got: %v", test.want, test.target)
			}
		})
	}
}

func TestDecodeJSONStrict(t *testing.T) {
	tests := []struct {
		name    string
		srcJSON string
		target  interface{}
		want    interface{}
		err     error
	}{
		{
			name:    "EmptyJSON",
			srcJSON: `{}`,
			target:  &struct{}{},
			want:    &struct{}{},
		},
		{
			name:    "IncorrectType",
			srcJSON: `{"number": "abc"}`,
			target:  &map[string]int{},
			err:     errors.New("body contains incorrect JSON type"),
		},
		{
			name:    "UnknownField",
			srcJSON: `{"unknown": 123}`,
			target:  &struct{ known string }{},
			err:     errors.New("body contains unknown key "),
		},
		{
			name:    "DisallowedField is OK in DecodeJSONStrict because it can figure out the type",
			srcJSON: `{"data": 123}`,
			target:  &struct{ Data interface{} }{},
			err:     nil,
			want:    &struct{ Data interface{} }{Data: float64(123)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.srcJSON))
			w := httptest.NewRecorder()
			err := httpgo.DecodeJSONStrict(w, r, test.target)
			if err != nil && test.err == nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil && test.err != nil {
				t.Errorf("expected error: %v", test.err)
			}
			if test.want != nil {
				assert.EqualValues(t, test.target, test.want, "expected value: %v, but got: %v", test.want, test.target)
			}
		})
	}
}
