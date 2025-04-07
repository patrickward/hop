package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/v2/route/middleware"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name          string
		options       func(*middleware.CORSOptions)
		method        string
		origin        string
		reqHeaders    map[string]string
		expectHeaders map[string]string
		expectStatus  int
	}{
		{
			name:    "default options",
			options: nil,
			method:  "GET",
			origin:  "*",
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "custom origins",
			options: func(opts *middleware.CORSOptions) {
				opts.AllowOrigins = []string{"https://example.com"}
			},
			method: "GET",
			origin: "https://example.com",
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin": "https://example.com",
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "preflight request",
			options: func(opts *middleware.CORSOptions) {
				opts.AllowOrigins = []string{"https://example.com"}
				opts.AllowMethods = []string{"GET", "POST"}
				opts.AllowHeaders = []string{"Content-Type"}
				opts.MaxAge = time.Hour
			},
			method: "OPTIONS",
			origin: "https://example.com",
			reqHeaders: map[string]string{
				"Access-Control-Request-Method":  "POST",
				"Access-Control-Request-Headers": "Content-Type",
			},
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET, POST",
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Max-Age":       "3600",
			},
			expectStatus: http.StatusNoContent,
		},
		{
			name: "credentials allowed",
			options: func(opts *middleware.CORSOptions) {
				opts.AllowOrigins = []string{"https://example.com"}
				opts.AllowCredentials = true
			},
			method: "GET",
			origin: "https://example.com",
			expectHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://example.com",
				"Access-Control-Allow-Credentials": "true",
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "origin not allowed",
			options: func(opts *middleware.CORSOptions) {
				opts.AllowOrigins = []string{"https://example.com"}
			},
			method:        "GET",
			origin:        "https://malicious.com",
			expectHeaders: map[string]string{},
			expectStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.CORS(tt.options)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "http://example.com", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			for k, v := range tt.reqHeaders {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectStatus, rec.Code)
			for k, v := range tt.expectHeaders {
				assert.Equal(t, v, rec.Header().Get(k))
			}
		})
	}
}
