package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/hop/route/middleware"
)

func TestPreventCSRF(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Setting up CSRF cookie
	r.AddCookie(&http.Cookie{
		Name:     "csrf",
		Value:    "_csrf_token_",
		HttpOnly: true,
	})

	handler := middleware.PreventCSRF(middleware.PreventCSRFOptions{
		HTTPOnly: true,
		Path:     "/",
		MaxAge:   86400,
		SameSite: "lax",
		Secure:   false,
	})

	handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	// CSRF cookie should be set
	cookie := w.Result().Cookies()[0]
	if cookie.Name != "csrf_token" {
		t.Errorf("Cookie name is not 'csrf'. Name is %s.", cookie.Name)
	}
	if cookie.HttpOnly != true {
		t.Errorf("Cookie HttpOnly is not true")
	}
	if cookie.Path != "/" {
		t.Errorf("Cookie Path is not '/'")
	}
	if cookie.MaxAge != 86400 {
		t.Errorf("Cookie MaxAge is not 86400")
	}
}
