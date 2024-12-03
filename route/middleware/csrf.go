package middleware

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// PreventCSRF prevents CSRF attacks by setting a CSRF cookie.
func PreventCSRF(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})

	return csrfHandler
}
