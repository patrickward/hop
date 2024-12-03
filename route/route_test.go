package route_test

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/route"
)

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
				m.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {})
				m.Post("/api/users", func(w http.ResponseWriter, r *http.Request) {})
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
				m.Get("/api/users", func(w http.ResponseWriter, r *http.Request) {
					_, err := w.Write([]byte("hello"))
					require.NoError(t, err)
				})
			},
			request:        httptest.NewRequest(http.MethodHead, "/api/users", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   "", // HEAD requests should have no body
		},
		{
			name: "Multiple methods on same path",
			setupRoutes: func(m *route.Mux) {
				m.Get("/api/resource", func(w http.ResponseWriter, r *http.Request) {})
				m.Post("/api/resource", func(w http.ResponseWriter, r *http.Request) {})
				m.Put("/api/resource", func(w http.ResponseWriter, r *http.Request) {})
				m.Delete("/api/resource", func(w http.ResponseWriter, r *http.Request) {})
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
				m.Get("/api/test", func(w http.ResponseWriter, r *http.Request) {
					_, err := w.Write([]byte("test"))
					require.NoError(t, err)
				})
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
				api := m.Group("/api", func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-API-Version", "1.0")
						next.ServeHTTP(w, r)
					})
				})

				v1 := api.Group("/v1")
				users := v1.Group("/users")

				users.Get("", func(w http.ResponseWriter, r *http.Request) {
					_, err := w.Write([]byte("users"))
					require.NoError(t, err)
				})
				users.Post("", func(w http.ResponseWriter, r *http.Request) {})
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
				api := m.Group("/api")
				v1 := api.Group("/v1")
				users := v1.Group("/users")

				users.Get("", func(w http.ResponseWriter, r *http.Request) {})
				users.Post("", func(w http.ResponseWriter, r *http.Request) {})
			},
			request:        httptest.NewRequest(http.MethodOptions, "/api/v1/users", nil),
			expectedStatus: http.StatusNoContent,
			expectedAllow:  []string{http.MethodGet, http.MethodHead, http.MethodPost},
		},
		{
			name: "Multiple nested groups with different methods",
			setupRoutes: func(m *route.Mux) {
				api := m.Group("/api")
				v1 := api.Group("/v1")

				// Register handlers at different levels
				api.Get("/health", func(w http.ResponseWriter, r *http.Request) {})
				v1.Get("/status", func(w http.ResponseWriter, r *http.Request) {})

				users := v1.Group("/users")
				users.Get("", func(w http.ResponseWriter, r *http.Request) {})
				users.Post("", func(w http.ResponseWriter, r *http.Request) {})
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

// TestListRoutes tests the ListRoutes functionality
func TestListRoutes(t *testing.T) {
	mux := route.New()

	// Setup some routes at different levels
	api := mux.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")

	api.Get("/health", func(w http.ResponseWriter, r *http.Request) {})
	v1.Get("/status", func(w http.ResponseWriter, r *http.Request) {})
	users.Get("", func(w http.ResponseWriter, r *http.Request) {})
	users.Post("", func(w http.ResponseWriter, r *http.Request) {})

	routes := mux.ListRoutes()

	// Create a map for easier testing
	routeMap := make(map[string][]string)
	for _, r := range routes {
		methods := make([]string, 0)
		for _, method := range r.Methods {
			methods = append(methods, method)
		}
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
	api := mux.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")

	api.Get("/health", func(w http.ResponseWriter, r *http.Request) {})
	v1.Get("/status", func(w http.ResponseWriter, r *http.Request) {})
	users.Get("", func(w http.ResponseWriter, r *http.Request) {})
	users.Post("", func(w http.ResponseWriter, r *http.Request) {})

	routesJSON, err := mux.DumpRoutes()
	require.NoError(t, err)

	// Expected JSON output
	expectedJSON := `[
	  {	
		"pattern": "/api/health",
		"methods": ["GET", "HEAD"] 
	  },
	  {
		"pattern": "/api/v1/status",
		"methods": ["GET", "HEAD"]
	  },
	  {
		"pattern": "/api/v1/users",
		"methods": ["GET", "HEAD", "POST"]
	  }
	]`

	assert.JSONEq(t, expectedJSON, routesJSON)
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
