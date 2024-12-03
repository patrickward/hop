package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

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

	middleware.PreventCSRF(fakeHandler()).ServeHTTP(w, r)
	assert.Equal(t, "fakeHandler", w.Body.String())
}

func fakeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "fakeHandler")
	})
}
