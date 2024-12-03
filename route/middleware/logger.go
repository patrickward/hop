package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Logger returns middleware that logs all requests using slog
//
// Example:
//
//	router.Use(middleware.Logger(logger, slog.Info))
//
// This will log all requests using the provided slog.Logger at the Info level.
func Logger(l *slog.Logger, level slog.Level) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r)

			l.LogAttrs(context.Background(), level, "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", rw.status),
				slog.Int64("bytes", rw.written),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
