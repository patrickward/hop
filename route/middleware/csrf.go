package middleware

import (
	"net/http"

	"github.com/justinas/nosurf"

	"github.com/patrickward/hop/v2/route"
	"github.com/patrickward/hop/v2/utils"
)

// PreventCSRFOptions provides options for PreventCSRF
type PreventCSRFOptions struct {
	HTTPOnly bool
	Path     string
	MaxAge   int
	SameSite string
	Secure   bool
}

// PreventCSRF prevents CSRF attacks by setting a CSRF cookie.
func PreventCSRF(opts PreventCSRFOptions) route.Middleware {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)

		sameSite := utils.SameSiteFromString(opts.SameSite)
		csrfHandler.SetBaseCookie(http.Cookie{
			HttpOnly: opts.HTTPOnly,
			Path:     opts.Path,
			MaxAge:   opts.MaxAge,
			SameSite: sameSite,
			Secure:   opts.Secure,
		})

		return csrfHandler
	}
}
