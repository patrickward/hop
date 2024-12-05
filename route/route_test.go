package route_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/route"
)

func emptyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func TestMux(t *testing.T) {
	tests := []struct {
		name           string
		setupRoutes    func(*route.Mux)
		request        *http.Request
		expectedStatus int
		expectedAllow  []string
		expectedBody   string
	}{
		{
			name: "OPTIONS request for registered path",
			setupRoutes: func(m *route.Mux) {
				m.Get("/api/users", emptyHandler())
				m.Post("/api/users", emptyHandler())
			},
			request:        httptest.NewRequest(http.MethodOptions, "/api/users", nil),
			expectedStatus: http.StatusNoContent,
			expectedAllow:  []string{http.MethodGet, http.MethodHead, http.MethodPost},
		},
		{
			name:           "OPTIONS request for unregistered path",
			setupRoutes:    func(m *route.Mux) {},
			request:        httptest.NewRequest(http.MethodOptions, "/api/notfound", nil),
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "HEAD request for GET route",
			setupRoutes: func(m *route.Mux) {
				m.Get("/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, err := w.Write([]byte("hello"))
					require.NoError(t, err)
				}))
			},
			request:        httptest.NewRequest(http.MethodHead, "/api/users", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   "", // HEAD requests should have no body
		},
		{
			name: "Multiple methods on same path",
			setupRoutes: func(m *route.Mux) {
				m.Get("/api/resource", emptyHandler())
				m.Post("/api/resource", emptyHandler())
				m.Put("/api/resource", emptyHandler())
				m.Delete("/api/resource", emptyHandler())
			},
			request:        httptest.NewRequest(http.MethodOptions, "/api/resource", nil),
			expectedStatus: http.StatusNoContent,
			expectedAllow:  []string{http.MethodDelete, http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut},
		},
		{
			name: "Middleware execution",
			setupRoutes: func(m *route.Mux) {
				m.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test", "middleware")
						next.ServeHTTP(w, r)
					})
				})
				m.Get("/api/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, err := w.Write([]byte("test"))
					require.NoError(t, err)
				}))
			},
			request:        httptest.NewRequest(http.MethodGet, "/api/test", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   "test",
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
		})
	}
}

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

// TestListRoutes tests the ListRoutes functionality
func TestListRoutes(t *testing.T) {
	mux := route.New()

	// Setup some routes at different levels
	mux.PrefixGroup("/api", func(group *route.Group) {
		group.Get("/health", emptyHandler())
		group.PrefixGroup("/v1", func(group *route.Group) {
			group.Get("/status", emptyHandler())
			group.PrefixGroup("/users", func(group *route.Group) {
				group.Get("", emptyHandler())
				group.Post("", emptyHandler())
			})
		})
	})

	routes := mux.ListRoutes()

	// Create a map for easier testing
	routeMap := make(map[string][]string)
	for _, r := range routes {
		methods := make([]string, 0)
		methods = append(methods, r.Methods...)
		routeMap[r.Pattern] = methods
	}

	expectedRoutes := map[string][]string{
		"/api/health":    {http.MethodGet, http.MethodHead},
		"/api/v1/status": {http.MethodGet, http.MethodHead},
		"/api/v1/users":  {http.MethodGet, http.MethodHead, http.MethodPost},
	}

	assert.Equal(t, len(expectedRoutes), len(routeMap), "Should have expected number of routes")

	for pattern, expectedMethods := range expectedRoutes {
		methods, exists := routeMap[pattern]
		require.True(t, exists, "Route %s should exist", pattern)
		assert.ElementsMatch(t, expectedMethods, methods, "Methods for route %s should match", pattern)
	}
}

// TestDumpRoutes tests the DumpRoutes functionality
func TestDumpRoutes(t *testing.T) {
	mux := route.New()

	// Setup some routes at different levels
	mux.PrefixGroup("/api", func(group *route.Group) {
		group.Get("/health", emptyHandler())
		group.PrefixGroup("/v1", func(group *route.Group) {
			group.Get("/status", emptyHandler())
			group.PrefixGroup("/users", func(group *route.Group) {
				group.Get("", emptyHandler())
				group.Post("", emptyHandler())
			})
		})
	})

	routesJSON, err := mux.DumpRoutes()
	require.NoError(t, err)

	expectedRoutes := []map[string]any{
		{
			"pattern": "/api/health",
			"methods": []any{"GET", "HEAD"},
		},
		{
			"pattern": "/api/v1/status",
			"methods": []any{"GET", "HEAD"},
		},
		{
			"pattern": "/api/v1/users",
			"methods": []any{"GET", "HEAD", "POST"},
		},
	}

	// Unmarshal and compare JSON
	var routes []map[string]any
	err = json.Unmarshal([]byte(routesJSON), &routes)
	require.NoError(t, err)

	// compare the two slices, regardless of order
	assert.ElementsMatch(t, expectedRoutes, routes)
}

// Helper function to parse Allow header
func parseAllowHeader(allow string) []string {
	if allow == "" {
		return nil
	}
	methods := strings.Split(allow, ", ")
	sort.Strings(methods)
	return methods
}
