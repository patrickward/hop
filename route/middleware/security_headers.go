package middleware

import (
	"net/http"

	"github.com/patrickward/hop/route"
)

// SecurityHeadersOptions contains configuration for security headers
type SecurityHeadersOptions struct {
	// ReferrerPolicy controls the Referrer-Policy header.
	// Default is "origin-when-cross-origin"
	ReferrerPolicy string

	// ContentTypeOptions controls the X-Content-Type-Options header.
	// Default is "nosniff"
	ContentTypeOptions string

	// FrameOptions controls the X-Frame-Options header.
	// Default is "deny"
	FrameOptions string

	// StrictTransportSecurity controls the Strict-Transport-Security header.
	// Empty string means the header won't be set
	StrictTransportSecurity string

	// Additional headers to set
	Additional map[string]string
}

// SecurityHeaders middleware sets security headers with configurable options
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers for more information
//
// Example (defaults):
//
//	router.Use(middleware.SecurityHeaders(nil))
//
// Example (custom):
//
//	router.Use(middleware.SecurityHeaders(func(opts *middleware.SecurityHeadersOptions) {
//		opts.ReferrerPolicy = "no-referrer"
//		opts.ContentTypeOptions = "nosniff"
//		opts.FrameOptions = "sameorigin"
//		opts.StrictTransportSecurity = "max-age=63072000; includeSubDomains"
//		opts.Additional["X-Custom-Header"] = "custom-value"
//	}))
func SecurityHeaders(optsFunc func(*SecurityHeadersOptions)) route.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Default options
			opts := SecurityHeadersOptions{
				ReferrerPolicy:     "origin-when-cross-origin",
				ContentTypeOptions: "nosniff",
				FrameOptions:       "deny",
				Additional:         make(map[string]string),
			}

			if optsFunc != nil {
				optsFunc(&opts)
			}

			// Set standard security headers
			if opts.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", opts.ReferrerPolicy)
			}
			if opts.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", opts.ContentTypeOptions)
			}
			if opts.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", opts.FrameOptions)
			}
			if opts.StrictTransportSecurity != "" {
				w.Header().Set("Strict-Transport-Security", opts.StrictTransportSecurity)
			}

			// Set additional headers
			for k, v := range opts.Additional {
				w.Header().Set(k, v)
			}

			next.ServeHTTP(w, r)
		})
	}
}
