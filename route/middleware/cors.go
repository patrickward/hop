package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CORSOptions contains the configuration for CORS middleware
type CORSOptions struct {
	// AllowOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// Default value is []string{"*"}
	AllowOrigins []string

	// AllowMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (GET, POST, HEAD).
	AllowMethods []string

	// AllowHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests. Default value is [].
	AllowHeaders []string

	// ExposeHeaders indicates which headers are safe to expose to the API of a
	// CORS response. Default value is [].
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	// Default value is false.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached. Default value is 12 hours.
	MaxAge time.Duration

	// OptionsSuccessStatus provides a status code to use for successful OPTIONS requests.
	// Default value is 204.
	OptionsSuccessStatus int
}

// CORS provides Cross-Origin Resource Sharing middleware
// For more information, see https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
//
// Example:
//
//	router.Use(middleware.CORS(func(opts *middleware.CORSOptions) {
//		opts.AllowOrigins = []string{"https://example.com"}
//	}))
func CORS(optsFunc func(opts *CORSOptions)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set up default options
			opts := CORSOptions{
				AllowOrigins:         []string{"*"},
				AllowMethods:         []string{"GET", "POST", "HEAD"},
				AllowHeaders:         []string{},
				ExposeHeaders:        []string{},
				AllowCredentials:     false,
				MaxAge:               12 * time.Hour,
				OptionsSuccessStatus: http.StatusNoContent,
			}

			// Apply custom options if provided
			if optsFunc != nil {
				optsFunc(&opts)
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				handlePreflight(w, r, &opts)
				return
			}

			// Handle actual requests
			handleActual(w, r, &opts)
			next.ServeHTTP(w, r)
		})
	}
}

func handlePreflight(w http.ResponseWriter, r *http.Request, opts *CORSOptions) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if !isOriginAllowed(origin, opts.AllowOrigins) {
		return
	}

	// Set CORS headers for preflight
	w.Header().Set("Access-Control-Allow-Origin", origin)
	if opts.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Handle Access-Control-Request-Method
	if requestMethod := r.Header.Get("Access-Control-Request-Method"); requestMethod != "" {
		if isMethodAllowed(requestMethod, opts.AllowMethods) {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(opts.AllowMethods, ", "))
		}
	}

	// Handle Access-Control-Request-Headers
	if requestHeaders := r.Header.Get("Access-Control-Request-Headers"); requestHeaders != "" {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(opts.AllowHeaders, ", "))
	}

	if len(opts.ExposeHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ", "))
	}

	if opts.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(opts.MaxAge.Seconds())))
	}

	w.WriteHeader(opts.OptionsSuccessStatus)
}

func handleActual(w http.ResponseWriter, r *http.Request, opts *CORSOptions) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if !isOriginAllowed(origin, opts.AllowOrigins) {
		return
	}

	// Set CORS headers for actual request
	w.Header().Set("Access-Control-Allow-Origin", origin)
	if opts.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if len(opts.ExposeHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(opts.ExposeHeaders, ", "))
	}
}

// Helper functions for CORS checks
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
	}

	return false
}

func isMethodAllowed(method string, allowedMethods []string) bool {
	if len(allowedMethods) == 0 {
		return false
	}

	method = strings.ToUpper(method)
	for _, allowedMethod := range allowedMethods {
		if strings.ToUpper(allowedMethod) == method {
			return true
		}
	}

	return false
}
