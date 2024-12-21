package pulse

import (
	"context"
	"crypto/subtle"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/patrickward/hop/route"
)

// Module implements hop.Module for metrics collection
type Module struct {
	collector Collector
	config    *Config
	ticker    *time.Ticker
	done      chan struct{}
}

type Config struct {
	// RoutePath is the endpoint where metrics are exposed
	RoutePath string
	// RouteUsername is the username required to access the metrics endpoint
	RouteUsername string
	// RoutePassword is the password required to access the metrics endpoint
	RoutePassword string
	// EnablePprof enables pprof endpoints
	EnablePprof bool
	// CollectionInterval is how often to collect system metrics
	CollectionInterval time.Duration
}

func NewModule(collector Collector, config *Config) *Module {
	if config == nil {
		config = &Config{
			RoutePath:          "/pulse",
			EnablePprof:        false,
			CollectionInterval: 15 * time.Second,
		}
	}

	if config.RoutePath == "" {
		config.RoutePath = "/pulse"
	}

	if config.CollectionInterval == 0 {
		config.CollectionInterval = 15 * time.Second
	}

	return &Module{
		collector: collector,
		config:    config,
		done:      make(chan struct{}),
	}
}

func (m *Module) ID() string {
	return "hop.pulse"
}

func (m *Module) Init() error {
	return nil
}

func (m *Module) RegisterRoutes(router *route.Mux) {
	// The middleware needs to be added at the top level to capture all requests
	router.Use(m.MetricsMiddleware())

	// Register metrics endpoint, use a group to apply auth middleware if configured
	router.Group(func(g *route.Group) {
		if m.config.RouteUsername != "" && m.config.RoutePassword != "" {
			g.Use(m.AuthMiddleware())
		}
		g.Get(m.config.RoutePath, m.collector.Handler())
	})

	// Optionally register pprof endpoints
	if m.config.EnablePprof {
		router.HandleFunc("/pulse/pprof/", http.HandlerFunc(pprof.Index))
		router.HandleFunc("/pulse/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		router.HandleFunc("/pulse/pprof/profile", http.HandlerFunc(pprof.Profile))
		router.HandleFunc("/pulse/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		router.HandleFunc("/pulse/pprof/trace", http.HandlerFunc(pprof.Trace))
	}
}

// Start begins periodic collection of system metrics
func (m *Module) Start(ctx context.Context) error {
	// Force initial collection
	m.collector.RecordMemStats()
	m.collector.RecordGoroutineCount()

	m.ticker = time.NewTicker(m.config.CollectionInterval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.done:
				return
			case <-m.ticker.C:
				m.collector.RecordMemStats()
				m.collector.RecordGoroutineCount()
			}
		}
	}()

	return nil
}

// Stop halts metric collection
func (m *Module) Stop(ctx context.Context) error {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	close(m.done)
	return nil
}

// MetricsMiddleware creates route.Middleware for collecting HTTP metrics
func (m *Module) MetricsMiddleware() route.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Track concurrent requests
			m.collector.IncrementConcurrentRequests()
			defer m.collector.DecrementConcurrentRequests()

			// Wrap response writer to capture status code
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			m.collector.RecordHTTPRequest(r.Method, r.URL.Path, duration, rw.statusCode)
		})
	}
}

// AuthMiddleware creates route.Middleware for authenticating requests to the metrics endpoint
func (m *Module) AuthMiddleware() route.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if m.config.RoutePassword == "" {
				next.ServeHTTP(w, r)
				return
			}

			username, password, ok := r.BasicAuth()
			if !ok {
				unauthorized(w)
				return
			}

			if !secureCompare(username, m.config.RouteUsername) ||
				!secureCompare(password, m.config.RoutePassword) {
				unauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Pulse Restricted"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// secureCompare performs a constant-time comparison of two strings to prevent timing attacks.
//
// This function is more secure than a direct subtle.ConstantTimeCompare because it first
// checks string lengths in constant time before comparing contents. This prevents timing
// attacks that could otherwise determine string length by measuring comparison duration.
//
// Security properties:
//   - Length check is performed in constant time using subtle.ConstantTimeEq
//   - Content comparison only occurs if lengths match
//   - All wrong-length attempts complete in identical time
//   - No timing information is leaked about partial matches
//
// Example timing attack prevented:
//
//	Real password: "secret123" (9 chars)
//	Without length check, comparing against:
//	  "a"         (1 char)  -> faster comparison
//	  "abcdefghi" (9 chars) -> slower comparison
//	With length check, all wrong-length attempts take identical time
//
// Returns true only if the strings are identical in both length and content.
func secureCompare(given, actual string) bool {
	// First check lengths in constant time
	if subtle.ConstantTimeEq(int32(len(given)), int32(len(actual))) == 1 {
		// Only if lengths match, check contents
		return subtle.ConstantTimeCompare([]byte(given), []byte(actual)) == 1
	}
	return false
}
