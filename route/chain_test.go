package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/v2/route"
)

func TestChain(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (http.Handler, *[]string)
		expectedOrder  []string
		expectedStatus int
	}{
		{
			name: "single middleware",
			setup: func() (http.Handler, *[]string) {
				order := make([]string, 0)
				middleware := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						order = append(order, "middleware")
						next.ServeHTTP(w, r)
					})
				}

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, "handler")
					w.WriteHeader(http.StatusOK)
				})

				chain := route.NewChain(middleware)
				return chain.Then(handler), &order
			},
			expectedOrder:  []string{"middleware", "handler"},
			expectedStatus: http.StatusOK,
		},
		{
			name: "multiple middleware execution order",
			setup: func() (http.Handler, *[]string) {
				order := make([]string, 0)
				middleware1 := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						order = append(order, "middleware1")
						next.ServeHTTP(w, r)
					})
				}
				middleware2 := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						order = append(order, "middleware2")
						next.ServeHTTP(w, r)
					})
				}

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, "handler")
					w.WriteHeader(http.StatusOK)
				})

				chain := route.NewChain(middleware1, middleware2)
				return chain.Then(handler), &order
			},
			expectedOrder:  []string{"middleware1", "middleware2", "handler"},
			expectedStatus: http.StatusOK,
		},
		{
			name: "chain append",
			setup: func() (http.Handler, *[]string) {
				order := make([]string, 0)
				middleware1 := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						order = append(order, "middleware1")
						next.ServeHTTP(w, r)
					})
				}
				middleware2 := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						order = append(order, "middleware2")
						next.ServeHTTP(w, r)
					})
				}

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, "handler")
					w.WriteHeader(http.StatusOK)
				})

				chain1 := route.NewChain(middleware1)
				chain2 := chain1.Append(middleware2)
				return chain2.Then(handler), &order
			},
			expectedOrder:  []string{"middleware1", "middleware2", "handler"},
			expectedStatus: http.StatusOK,
		},
		{
			name: "chain extend",
			setup: func() (http.Handler, *[]string) {
				order := make([]string, 0)
				middleware1 := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						order = append(order, "middleware1")
						next.ServeHTTP(w, r)
					})
				}
				middleware2 := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						order = append(order, "middleware2")
						next.ServeHTTP(w, r)
					})
				}

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, "handler")
					w.WriteHeader(http.StatusOK)
				})

				chain1 := route.NewChain(middleware1)
				chain2 := route.NewChain(middleware2)
				chain3 := chain1.Extend(chain2)
				return chain3.Then(handler), &order
			},
			expectedOrder:  []string{"middleware1", "middleware2", "handler"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, order := tt.setup()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedOrder, *order)
		})
	}
}

func TestBeforeAfter(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() (http.Handler, *[]string)
		expectedOrder []string
	}{
		{
			name: "before middleware",
			setup: func() (http.Handler, *[]string) {
				order := make([]string, 0)
				middleware := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						order = append(order, "before")
						next.ServeHTTP(w, r)
					})
				}

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, "handler")
				})

				return route.Before(handler, middleware), &order
			},
			expectedOrder: []string{"before", "handler"},
		},
		{
			name: "after middleware",
			setup: func() (http.Handler, *[]string) {
				order := make([]string, 0)
				middleware := func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						next.ServeHTTP(w, r)
						order = append(order, "after")
					})
				}

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					order = append(order, "handler")
				})

				return route.After(handler, middleware), &order
			},
			expectedOrder: []string{"handler", "after"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, order := tt.setup()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedOrder, *order)
		})
	}
}
