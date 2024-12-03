package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Timeout returns middleware that cancels requests after a timeout
//
// Example:
//
//	router.Use(middleware.Timeout(5 * time.Second))
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					http.Error(w, "Request Timeout", http.StatusGatewayTimeout)
				}
			}
		})
	}
}
