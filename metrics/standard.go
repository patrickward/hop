package metrics

import (
	"expvar"
	"runtime"
	"sync"
	"time"
)

// StandardCollector implements Collector using the standard library
type StandardCollector struct {
	mu         sync.RWMutex
	serverName string
	startTime  time.Time
	counters   map[string]*standardCounter
	gauges     map[string]*standardGauge
	histograms map[string]*standardHistogram

	// Pre-allocated metrics for performance
	httpRequests  *standardCounter
	httpDurations *standardHistogram
	httpErrors    *standardCounter

	// System metrics
	goroutines *standardGauge
	memAlloc   *standardGauge
	memTotal   *standardGauge
	memSys     *standardGauge
	gcPauses   *standardHistogram
}

// StandardCollectorOption is a functional option for configuring a StandardCollector
type StandardCollectorOption func(*StandardCollector)

// WithServerName sets the server name for the collector
func WithServerName(name string) StandardCollectorOption {
	return func(c *StandardCollector) {
		c.serverName = name
	}
}

// NewStandardCollector creates a new StandardCollector
func NewStandardCollector(opts ...StandardCollectorOption) *StandardCollector {
	c := &StandardCollector{
		serverName: "HOP Server",
		startTime:  time.Now(),
		counters:   make(map[string]*standardCounter),
		gauges:     make(map[string]*standardGauge),
		histograms: make(map[string]*standardHistogram),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Initialize common metrics
	c.httpRequests = c.getOrCreateCounter("http_requests_total")
	c.httpDurations = c.getOrCreateHistogram("http_request_duration_ms")
	c.httpErrors = c.getOrCreateCounter("http_errors_total")

	c.goroutines = c.getOrCreateGauge("goroutines")
	c.memAlloc = c.getOrCreateGauge("memory_alloc_bytes")
	c.memTotal = c.getOrCreateGauge("memory_total_bytes")
	c.memSys = c.getOrCreateGauge("memory_sys_bytes")
	c.gcPauses = c.getOrCreateHistogram("gc_pause_ms")

	return c
}

// Counter implementation
type standardCounter struct {
	v *expvar.Int
}

// Inc increments the counter by 1
func (c *standardCounter) Inc() { c.v.Add(1) }

// Add increments the counter by the given delta
func (c *standardCounter) Add(delta float64) { c.v.Add(int64(delta)) }

// Value returns the current value of the counter
func (c *standardCounter) Value() float64 { return float64(c.v.Value()) }

// Gauge implementation
type standardGauge struct {
	v *expvar.Float
}

// Set sets the gauge to the given value
func (g *standardGauge) Set(value float64) { g.v.Set(value) }

// Add increments the gauge by the given delta
func (g *standardGauge) Add(delta float64) { g.v.Add(delta) }

// Sub decrements the gauge by the given delta
func (g *standardGauge) Sub(delta float64) { g.v.Add(-delta) }

// Value returns the current value of the gauge
func (g *standardGauge) Value() float64 { return g.v.Value() }

// Histogram implementation using expvar
type standardHistogram struct {
	mu      sync.RWMutex
	count   uint64
	sum     float64
	buckets map[float64]uint64
}

// Observe records a new observation
func (h *standardHistogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.count++
	h.sum += value
	for bound := range h.buckets {
		if value <= bound {
			h.buckets[bound]++
		}
	}
}

// Count returns the number of observations
func (h *standardHistogram) Count() uint64 { return h.count }

// Sum returns the sum of all observations
func (h *standardHistogram) Sum() float64 { return h.sum }

// Counter returns a counter metric
func (c *StandardCollector) Counter(name string) Counter {
	return c.getOrCreateCounter(name)
}

// Gauge returns a gauge metric
func (c *StandardCollector) Gauge(name string) Gauge {
	return c.getOrCreateGauge(name)
}

// Histogram returns a histogram metric
func (c *StandardCollector) Histogram(name string) Histogram {
	return c.getOrCreateHistogram(name)
}

// RecordMemStats captures memory statistics
func (c *StandardCollector) RecordMemStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	c.memAlloc.Set(float64(ms.Alloc))
	c.memTotal.Set(float64(ms.TotalAlloc))
	c.memSys.Set(float64(ms.Sys))
	c.gcPauses.Observe(float64(ms.PauseNs[(ms.NumGC+255)%256]) / 1e6) // Convert to milliseconds
}

// RecordGoroutineCount captures the number of goroutines
func (c *StandardCollector) RecordGoroutineCount() {
	c.goroutines.Set(float64(runtime.NumGoroutine()))
}

// RecordHTTPRequest records metrics about an HTTP request
func (c *StandardCollector) RecordHTTPRequest(method, path string, duration time.Duration, statusCode int) {
	c.httpRequests.Inc()
	c.httpDurations.Observe(float64(duration.Milliseconds()))

	if statusCode >= 400 {
		c.httpErrors.Inc()
	}
}

// Helper methods for creating metrics
func (c *StandardCollector) getOrCreateCounter(name string) *standardCounter {
	c.mu.Lock()
	defer c.mu.Unlock()

	if counter, exists := c.counters[name]; exists {
		return counter
	}

	counter := &standardCounter{v: expvar.NewInt(name)}
	c.counters[name] = counter
	return counter
}

// Helper methods for creating metrics

func (c *StandardCollector) getOrCreateGauge(name string) *standardGauge {
	c.mu.Lock()
	defer c.mu.Unlock()

	if gauge, exists := c.gauges[name]; exists {
		return gauge
	}

	gauge := &standardGauge{v: expvar.NewFloat(name)}
	c.gauges[name] = gauge
	return gauge
}

func (c *StandardCollector) getOrCreateHistogram(name string) *standardHistogram {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hist, exists := c.histograms[name]; exists {
		return hist
	}

	// Default buckets for latency-style metrics
	hist := &standardHistogram{
		buckets: map[float64]uint64{
			10:    0, // 10ms
			50:    0, // 50ms
			100:   0, // 100ms
			250:   0, // 250ms
			500:   0, // 500ms
			1000:  0, // 1s
			2500:  0, // 2.5s
			5000:  0, // 5s
			10000: 0, // 10s
		},
	}
	c.histograms[name] = hist

	// Register with expvar for exposure
	expvar.Publish(name, expvar.Func(func() interface{} {
		return map[string]interface{}{
			"count":   hist.Count(),
			"sum":     hist.Sum(),
			"buckets": hist.buckets,
		}
	}))

	return hist
}
