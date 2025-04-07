package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/hop/v2/route/middleware"
)

func TestSecurityHeaders_DefaultOptions(t *testing.T) {
	handler := middleware.SecurityHeaders(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Referrer-Policy"); got != "origin-when-cross-origin" {
		t.Errorf("Referrer-Policy = %v, want %v", got, "origin-when-cross-origin")
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %v, want %v", got, "nosniff")
	}
	if got := w.Header().Get("X-Frame-Options"); got != "deny" {
		t.Errorf("X-Frame-Options = %v, want %v", got, "deny")
	}
}

func TestSecurityHeaders_CustomOptions(t *testing.T) {
	optsFunc := func(opts *middleware.SecurityHeadersOptions) {
		opts.ReferrerPolicy = "no-referrer"
		opts.ContentTypeOptions = "nosniff"
		opts.FrameOptions = "sameorigin"
		opts.StrictTransportSecurity = "max-age=63072000; includeSubDomains"
		opts.Additional["X-Custom-Header"] = "custom-value"
	}
	handler := middleware.SecurityHeaders(optsFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Errorf("Referrer-Policy = %v, want %v", got, "no-referrer")
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %v, want %v", got, "nosniff")
	}
	if got := w.Header().Get("X-Frame-Options"); got != "sameorigin" {
		t.Errorf("X-Frame-Options = %v, want %v", got, "sameorigin")
	}
	if got := w.Header().Get("Strict-Transport-Security"); got != "max-age=63072000; includeSubDomains" {
		t.Errorf("Strict-Transport-Security = %v, want %v", got, "max-age=63072000; includeSubDomains")
	}
	if got := w.Header().Get("X-Custom-Header"); got != "custom-value" {
		t.Errorf("X-Custom-Header = %v, want %v", got, "custom-value")
	}
}

func TestSecurityHeaders_EmptyOptions(t *testing.T) {
	optsFunc := func(opts *middleware.SecurityHeadersOptions) {
		opts.ReferrerPolicy = ""
		opts.ContentTypeOptions = ""
		opts.FrameOptions = ""
		opts.StrictTransportSecurity = ""
	}
	handler := middleware.SecurityHeaders(optsFunc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Referrer-Policy"); got != "" {
		t.Errorf("Referrer-Policy = %v, want empty", got)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "" {
		t.Errorf("X-Content-Type-Options = %v, want empty", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "" {
		t.Errorf("X-Frame-Options = %v, want empty", got)
	}
	if got := w.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("Strict-Transport-Security = %v, want empty", got)
	}
}
