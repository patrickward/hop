package metrics

import (
	"context"
	"fmt"
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
	// Path where metrics are exposed
	MetricsPath string
	// Enable pprof endpoints
	EnablePprof bool
	// How often to collect system metrics
	CollectionInterval time.Duration
}

func NewModule(collector Collector, config *Config) *Module {
	if config == nil {
		config = &Config{
			MetricsPath:        "/metrics",
			EnablePprof:        false,
			CollectionInterval: 15 * time.Second,
		}
	}

	if config.MetricsPath == "" {
		config.MetricsPath = "/metrics"
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
	return "hop.metrics"
}

func (m *Module) Init() error {
	return nil
}

func (m *Module) RegisterRoutes(router *route.Mux) {
	// Register metrics endpoint
	fmt.Println("Registering metrics endpoint: ", m.config.MetricsPath)
	router.Get(m.config.MetricsPath, m.collector.Handler())

	// Optionally register pprof endpoints
	if m.config.EnablePprof {
		router.HandleFunc("/debug/pprof/", http.HandlerFunc(pprof.Index))
		router.HandleFunc("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		router.HandleFunc("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		router.HandleFunc("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		router.HandleFunc("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
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

// Middleware creates route.Middleware for collecting HTTP metrics
func (m *Module) Middleware() route.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			m.collector.RecordHTTPRequest(r.Method, r.URL.Path, duration, rw.statusCode)
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
