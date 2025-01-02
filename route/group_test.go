package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/route"
)

func TestGroup(t *testing.T) {
	tests := []struct {
		name           string
		setupRoutes    func(*route.Mux)
		request        *http.Request
		expectedStatus int
		expectedAllow  []string
		expectedBody   string
		expectedHeader map[string]string
	}{
		{
			name: "Nested groups with middleware",
			setupRoutes: func(m *route.Mux) {
				m.PrefixGroup("/api", func(group *route.Group) {
					group.Use(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("X-API-Version", "1.0")
							next.ServeHTTP(w, r)
						})
					})

					group.PrefixGroup("/v1", func(group *route.Group) {
						group.PrefixGroup("/users", func(group *route.Group) {
							group.Get("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								_, err := w.Write([]byte("users"))
								require.NoError(t, err)
							}))
							group.Post("", emptyHandler())
						})
					})
				})
			},
			request:        httptest.NewRequest(http.MethodGet, "/api/v1/users", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   "users",
			expectedHeader: map[string]string{
				"X-API-Version": "1.0",
			},
		},
		{
			name: "OPTIONS for nested group route",
			setupRoutes: func(m *route.Mux) {
				m.PrefixGroup("/api", func(group *route.Group) {
					group.PrefixGroup("/v1", func(group *route.Group) {
						group.PrefixGroup("/users", func(group *route.Group) {
							group.Get("", emptyHandler())
							group.Post("", emptyHandler())
						})
					})
				})
			},
			request:        httptest.NewRequest(http.MethodOptions, "/api/v1/users", nil),
			expectedStatus: http.StatusNoContent,
			expectedAllow:  []string{http.MethodGet, http.MethodHead, http.MethodPost},
		},
		{
			name: "Multiple nested groups with different methods",
			setupRoutes: func(m *route.Mux) {
				m.PrefixGroup("/api", func(group *route.Group) {
					group.Get("/health", emptyHandler())
					group.PrefixGroup("/v1", func(group *route.Group) {
						group.Get("/status", emptyHandler())
						group.PrefixGroup("/users", func(group *route.Group) {
							group.Get("", emptyHandler())
							group.Post("", emptyHandler())
						})
					})
				})
			},
			request:        httptest.NewRequest(http.MethodOptions, "/api/v1/users", nil),
			expectedStatus: http.StatusNoContent,
			expectedAllow:  []string{http.MethodGet, http.MethodHead, http.MethodPost},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := route.New()
			tt.setupRoutes(mux)

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, tt.request)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if len(tt.expectedAllow) > 0 {
				allow := w.Header().Get("Allow")
				methods := parseAllowHeader(allow)
				assert.ElementsMatch(t, tt.expectedAllow, methods)
			}

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			for k, v := range tt.expectedHeader {
				assert.Equal(t, v, w.Header().Get(k))
			}
		})
	}
}

// TestGroupWithParentMiddleware tests that middleware is applied in the correct order
func TestGroupWithParentMiddleware(t *testing.T) {
	mux := route.New()

	// Add middleware to the root mux
	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Root", "1.0")
			next.ServeHTTP(w, r)
		})
	})

	// Add a group with its own middleware
	api := mux.PrefixGroup("/api", func(group *route.Group) {
		group.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-API", "1.0")
				next.ServeHTTP(w, r)
			})
		})
		group.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("users"))
			require.NoError(t, err)
		}))
	})

	// Add a nested group with its own middleware
	api.PrefixGroup("/v1", func(group *route.Group) {
		group.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-V1", "1.0")
				next.ServeHTTP(w, r)
			})
		})

		// Add a handler to the nested group
		group.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("users v1"))
			require.NoError(t, err)
		}))
	})

	// Make a request to the handler
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	mux.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "users", w.Body.String())
	assert.Equal(t, "1.0", w.Header().Get("X-Root"))
	assert.Equal(t, "1.0", w.Header().Get("X-API"))

	// Make a request to the nested handler
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	mux.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "users v1", w.Body.String())
	assert.Equal(t, "1.0", w.Header().Get("X-Root"))
	assert.Equal(t, "1.0", w.Header().Get("X-API"))
	assert.Equal(t, "1.0", w.Header().Get("X-V1"))
}

func TestIndependentGroups(t *testing.T) {
	tests := []struct {
		name           string
		setupRoutes    func(*route.Mux, *map[string]bool)
		request        string
		expectedCalled map[string]bool
		expectedStatus int
	}{
		{
			name: "independent group ignores mux middleware",
			setupRoutes: func(m *route.Mux, called *map[string]bool) {
				// Add mux-level middleware
				m.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						(*called)["mux"] = true
						next.ServeHTTP(w, r)
					})
				})

				// Add independent group with its own middleware
				m.PrefixGroup("/api", func(g *route.Group) {
					g.Independent()
					g.Use(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							(*called)["independent"] = true
							next.ServeHTTP(w, r)
						})
					})

					g.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						(*called)["handler"] = true
						w.WriteHeader(http.StatusOK)
					}))
				})
			},
			request: "/api/test",
			expectedCalled: map[string]bool{
				"independent": true,
				"handler":     true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "regular group inherits mux middleware",
			setupRoutes: func(m *route.Mux, called *map[string]bool) {
				// Add mux-level middleware
				m.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						(*called)["mux"] = true
						next.ServeHTTP(w, r)
					})
				})

				// Add regular group
				m.PrefixGroup("/api", func(g *route.Group) {
					g.Use(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							(*called)["group"] = true
							next.ServeHTTP(w, r)
						})
					})

					g.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						(*called)["handler"] = true
						w.WriteHeader(http.StatusOK)
					}))
				})
			},
			request: "/api/test",
			expectedCalled: map[string]bool{
				"mux":     true,
				"group":   true,
				"handler": true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "nested independent group",
			setupRoutes: func(m *route.Mux, called *map[string]bool) {
				// Add mux-level middleware
				m.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						(*called)["mux"] = true
						next.ServeHTTP(w, r)
					})
				})

				m.PrefixGroup("/api", func(g *route.Group) {
					g.Use(func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							(*called)["parent"] = true
							next.ServeHTTP(w, r)
						})
					})

					g.PrefixGroup("/webhooks", func(g *route.Group) {
						g.Independent()
						g.Use(func(next http.Handler) http.Handler {
							return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
								(*called)["webhook"] = true
								next.ServeHTTP(w, r)
							})
						})

						g.Get("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							(*called)["handler"] = true
							w.WriteHeader(http.StatusOK)
						}))
					})
				})
			},
			request: "/api/webhooks/test",
			expectedCalled: map[string]bool{
				"webhook": true,
				"handler": true,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := route.New()
			called := make(map[string]bool)

			tt.setupRoutes(mux, &called)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.request, nil)
			mux.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedCalled, called)
		})
	}
}
