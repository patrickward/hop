package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickward/hop/v2/route"
)

// noopHandler is a simple handler that does nothing
var noopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

// noopMiddleware is a middleware that just calls next
func noopMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// logMiddleware simulates a typical logging middleware
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some pre-processing
		path := r.URL.Path
		method := r.Method
		_ = path
		_ = method

		next.ServeHTTP(w, r)

		// Simulate some post-processing
		status := 200
		_ = status
	})
}

func BenchmarkChain(b *testing.B) {
	benchmarks := []struct {
		name  string
		setup func() http.Handler
	}{
		{
			name: "single_middleware",
			setup: func() http.Handler {
				chain := route.NewChain(noopMiddleware)
				return chain.Then(noopHandler)
			},
		},
		{
			name: "five_middlewares",
			setup: func() http.Handler {
				chain := route.NewChain(
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
				)
				return chain.Then(noopHandler)
			},
		},
		{
			name: "ten_middlewares",
			setup: func() http.Handler {
				chain := route.NewChain(
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
					noopMiddleware,
				)
				return chain.Then(noopHandler)
			},
		},
		{
			name: "practical_middleware_stack",
			setup: func() http.Handler {
				chain := route.NewChain(
					logMiddleware,  // logging
					noopMiddleware, // auth
					noopMiddleware, // metrics
					noopMiddleware, // tracing
					logMiddleware,  // request ID
				)
				return chain.Then(noopHandler)
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	for _, bm := range benchmarks {
		h := bm.setup()
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				h.ServeHTTP(w, req)
			}
		})
	}
}

func BenchmarkMux(b *testing.B) {
	benchmarks := []struct {
		name  string
		setup func() (*route.Mux, *http.Request)
	}{
		{
			name: "simple_route",
			setup: func() (*route.Mux, *http.Request) {
				mux := route.New()
				mux.Get("/api/users", noopHandler)
				return mux, httptest.NewRequest(http.MethodGet, "/api/users", nil)
			},
		},
		{
			name: "route_with_middleware",
			setup: func() (*route.Mux, *http.Request) {
				mux := route.New()
				mux.Use(logMiddleware)
				mux.Get("/api/users", noopHandler)
				return mux, httptest.NewRequest(http.MethodGet, "/api/users", nil)
			},
		},
		{
			name: "grouped_routes",
			setup: func() (*route.Mux, *http.Request) {
				mux := route.New()
				mux.PrefixGroup("/api", func(group *route.Group) {
					group.Get("/v1/users", noopHandler)
				})

				return mux, httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
			},
		},
		{
			name: "grouped_routes_with_middleware",
			setup: func() (*route.Mux, *http.Request) {
				mux := route.New()
				mux.Use(logMiddleware) // global middleware
				mux.PrefixGroup("/api", func(group *route.Group) {
					group.Get("/v1/users", noopHandler)
				})
				return mux, httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
			},
		},
		{
			name: "options_request",
			setup: func() (*route.Mux, *http.Request) {
				mux := route.New()
				mux.Get("/api/users", noopHandler)
				mux.Post("/api/users", noopHandler)
				mux.Put("/api/users", noopHandler)
				return mux, httptest.NewRequest(http.MethodOptions, "/api/users", nil)
			},
		},
	}

	w := httptest.NewRecorder()

	for _, bm := range benchmarks {
		mux, req := bm.setup()
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				mux.ServeHTTP(w, req)
			}
		})
	}
}

func BenchmarkChainOperations(b *testing.B) {
	benchmarks := []struct {
		name      string
		operation func()
	}{
		{
			name: "new_chain",
			operation: func() {
				_ = route.NewChain(noopMiddleware, noopMiddleware, noopMiddleware)
			},
		},
		{
			name: "chain_append",
			operation: func() {
				base := route.NewChain(noopMiddleware, noopMiddleware)
				_ = base.Append(noopMiddleware, noopMiddleware)
			},
		},
		{
			name: "chain_extend",
			operation: func() {
				base := route.NewChain(noopMiddleware, noopMiddleware)
				other := route.NewChain(noopMiddleware, noopMiddleware)
				_ = base.Extend(other)
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bm.operation()
			}
		})
	}
}

func BenchmarkRouteRegistration(b *testing.B) {
	benchmarks := []struct {
		name      string
		operation func()
	}{
		{
			name: "simple_route",
			operation: func() {
				mux := route.New()
				mux.Get("/api/users", noopHandler)
			},
		},
		{
			name: "route_with_middleware",
			operation: func() {
				mux := route.New()
				mux.Use(logMiddleware)
				mux.Get("/api/users", noopHandler)
			},
		},
		{
			name: "grouped_route",
			operation: func() {
				mux := route.New()
				//api := mux.PrefixGroup("/api")
				//v1 := api.PrefixGroup("/v1")
				//v1.Get("/users", noopHandler)

				mux.PrefixGroup("/api", func(group *route.Group) {
					group.PrefixGroup("/v1", func(group *route.Group) {
						group.Get("/users", noopHandler)
					})
				})
			},
		},
		{
			name: "multiple_methods_same_path",
			operation: func() {
				mux := route.New()
				mux.Get("/api/users", noopHandler)
				mux.Post("/api/users", noopHandler)
				mux.Put("/api/users", noopHandler)
				mux.Delete("/api/users", noopHandler)
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bm.operation()
			}
		})
	}
}
