package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/v2/route/middleware"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name           string
		handler        middleware.ErrorHandler
		panicValue     any
		expectStatus   int
		expectResponse string
	}{
		{
			name:           "no panic",
			handler:        nil,
			panicValue:     nil,
			expectStatus:   http.StatusOK,
			expectResponse: "OK",
		},
		{
			name:           "panic with default handler",
			handler:        nil,
			panicValue:     "something went wrong",
			expectStatus:   http.StatusInternalServerError,
			expectResponse: http.StatusText(http.StatusInternalServerError),
		},
		{
			name: "panic with custom handler",
			handler: func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, "custom error", http.StatusTeapot)
			},
			panicValue:     "something went wrong",
			expectStatus:   http.StatusTeapot,
			expectResponse: "custom error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
			handler := middleware.Recovery(logger, tt.handler)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.panicValue != nil {
					panic(tt.panicValue)
				}
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("OK"))
				assert.NoError(t, err)
			}))

			req := httptest.NewRequest("GET", "http://example.com", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectStatus, rec.Code)
			assert.Equal(t, tt.expectResponse, strings.TrimSpace(rec.Body.String()))
		})
	}
}
