package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/route/middleware"
)

func TestContentSecurityPolicy(t *testing.T) {
	tests := []struct {
		name             string
		options          func(*middleware.ContentSecurityPolicyOptions)
		expectDirectives []string
	}{
		{
			name:    "default CSP",
			options: nil,
			expectDirectives: []string{
				"default-src 'none'",
				"font-src 'self'",
				"img-src 'self'",
				"script-src 'self'",
				"style-src 'self'",
			},
		},
		{
			name: "custom CSP",
			options: func(opts *middleware.ContentSecurityPolicyOptions) {
				opts.DefaultSrc = "'self'"
				opts.ScriptSrc = "'self' https://apis.google.com"
				opts.StyleSrc = "'self' 'unsafe-inline'"
				opts.ImgSrc = "'self' data: https:"
				opts.ConnectSrc = "'self' https://api.example.com"
			},
			expectDirectives: []string{
				"font-src 'self'",
				"default-src 'self'",
				"script-src 'self' https://apis.google.com",
				"style-src 'self' 'unsafe-inline'",
				"img-src 'self' data: https:",
				"connect-src 'self' https://api.example.com",
			},
		},
		{
			name: "with report URI",
			options: func(opts *middleware.ContentSecurityPolicyOptions) {
				opts.DefaultSrc = "'self'"
				opts.ReportTo = "/csp-report"
			},
			expectDirectives: []string{
				"font-src 'self'",
				"default-src 'self'",
				"img-src 'self'",
				"script-src 'self'",
				"style-src 'self'",
				"report-to /csp-report",
			},
		},
		{
			name: "strict CSP",
			options: func(opts *middleware.ContentSecurityPolicyOptions) {
				opts.DefaultSrc = "'none'"
				opts.ScriptSrc = "'self'"
				opts.StyleSrc = "'self'"
				opts.ImgSrc = "'self'"
				opts.ConnectSrc = "'self'"
				opts.FormAction = "'self'"
				opts.FrameAncestors = "'none'"
				opts.BaseURI = "'self'"
				opts.ObjectSrc = "'none'"
			},
			expectDirectives: []string{
				"font-src 'self'",
				"default-src 'none'",
				"script-src 'self'",
				"style-src 'self'",
				"img-src 'self'",
				"connect-src 'self'",
				"form-action 'self'",
				"frame-ancestors 'none'",
				"base-uri 'self'",
				"object-src 'none'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.ContentSecurityPolicy(tt.options)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "http://example.com", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			gotPolicy := rec.Header().Get("Content-Security-Policy")
			gotDirectives := strings.Split(gotPolicy, ";")

			// Clean up the directives
			for i := range gotDirectives {
				gotDirectives[i] = strings.TrimSpace(gotDirectives[i])
			}

			// Compare directives regardless of order
			assert.ElementsMatch(t, tt.expectDirectives, gotDirectives)
		})
	}
}
