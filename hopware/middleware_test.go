package hopware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/hopware"
)

func fakeRender(w http.ResponseWriter, _ *http.Request, err error) {
	fmt.Fprintf(w, "System Error: %s", err.Error())
}

func fakeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "fakeHandler")
	})
}

func fakePanicHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("panic fakePanicHandler")
	})
}

func TestRecoverPanic(t *testing.T) {
	settings := []struct {
		Name           string
		handler        http.Handler
		render         func(http.ResponseWriter, *http.Request, error)
		expectedOutput string
	}{
		{
			Name:           "test fakeHandler with recover",
			handler:        fakeHandler(),
			render:         fakeRender,
			expectedOutput: "fakeHandler",
		},
		{
			Name:           "test fakeHandlerPanic with recover",
			handler:        fakePanicHandler(),
			render:         fakeRender,
			expectedOutput: "System Error: panic fakePanicHandler",
		},
	}

	for _, setting := range settings {
		t.Run(setting.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				hopware.RecoverPanic(setting.render)(setting.handler).ServeHTTP(w, r)
			})

			assert.Equal(t, setting.expectedOutput, w.Body.String())
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	hopware.SecurityHeaders(fakeHandler()).ServeHTTP(w, r)

	assert.Equal(t, "origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "deny", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "fakeHandler", w.Body.String())
}

func TestContentSecurityPolicy(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	hopware.ContentSecurityPolicy(nil)(fakeHandler()).ServeHTTP(w, r)

	policy := "default-src 'none';font-src 'self';img-src 'self';script-src 'self';style-src 'self'"
	assert.Equal(t, policy, w.Header().Get("Content-Security-Policy"))
}

func TestPreventCSRF(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Setting up CSRF cookie
	r.AddCookie(&http.Cookie{
		Name:     "csrf",
		Value:    "_csrf_token_",
		HttpOnly: true,
	})

	hopware.PreventCSRF(fakeHandler()).ServeHTTP(w, r)
	assert.Equal(t, "fakeHandler", w.Body.String())
}
