package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// ErrorHandler is a function that handles errors during request processing
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// Recovery returns middleware that recovers from panics and calls the optional error handler
// If no error handler is provided, a default error response is sent
//
// Example:
//
//	router.Use(middleware.Recovery(logger, func(w http.ResponseWriter, r *http.Request, err any) {
//		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
//	}))
func Recovery(logger *slog.Logger, handler ErrorHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					logger.Error("panic recovered",
						"error", err,
						"stack", string(stack),
					)

					if handler != nil {
						handler(w, r, fmt.Errorf("%v", err))
						return
					}

					// Default error handling
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
