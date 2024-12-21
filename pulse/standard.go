package pulse

import (
	"expvar"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ThresholdLevel is an enumeration of threshold levels
type ThresholdLevel int

const (
	// ThresholdInfo indicates a threshold is informational
	ThresholdInfo ThresholdLevel = iota
	// ThresholdOK indicates a threshold is within acceptable limits
	ThresholdOK
	// ThresholdWarning indicates a threshold is approaching a warning level
	ThresholdWarning
	// ThresholdCritical indicates a threshold has exceeded a critical level
	ThresholdCritical
)

// Thresholds configuration
type Thresholds struct {
	CPUPercent              float64 // CPU usage percentage
	ClientErrorRatePercent  float64 // Higher threshold for 4xx errors
	DiskPercent             float64 // Percentage of disk space used
	GCPauseMs               float64 // Warning when GC pauses exceed this duration
	GoroutineCount          int     // Number of goroutines
	MaxGCFrequency          float64 // Warning when GC runs too frequently (times per minute)
	MemoryGrowthRatePercent float64 // Warning when memory grows too fast (percent per minute)
	MemoryPercent           float64 // Percentage of total memory used
	ServerErrorRatePercent  float64 // Lower threshold for 5xx errors
}

// DefaultThresholds provides default threshold values
var DefaultThresholds = Thresholds{
	CPUPercent:              75.0,  // Warning at 75% CPU usage
	ClientErrorRatePercent:  40.0,  // Allow higher rate for client errors
	DiskPercent:             85.0,  // Warning at 85% disk usage
	GCPauseMs:               100.0, // 100ms pause time might affect responsiveness
	GoroutineCount:          1000,  // Warning at 1000 goroutines
	MaxGCFrequency:          100.0, // More than 100 GCs per minute might indicate pressure
	MemoryGrowthRatePercent: 20.0,  // 20% growth per minute might indicate a leak
	MemoryPercent:           80.0,  // Warning at 80% memory usage
	ServerErrorRatePercent:  1.0,   // Very low tolerance for server errors
}

// StandardCollector implements Collector using the standard library
type StandardCollector struct {
	mu         sync.RWMutex
	serverName string
	startTime  time.Time
	counters   map[string]*standardCounter
	gauges     map[string]*standardGauge
	histograms map[string]*standardHistogram

	// Pre-allocated metrics for performance
	httpRequests     *standardCounter
	httpDurations    *standardHistogram
	httpServerErrors *standardCounter
	httpClientErrors *standardCounter

	// System metrics
	goroutines *standardGauge
	//memAlloc   *standardGauge
	//memTotal   *standardGauge
	//memSys     *standardGauge
	gcPauses *standardHistogram

	// Enhanced memory metrics
	heapInuse    *standardGauge
	heapSys      *standardGauge
	heapIdle     *standardGauge
	heapReleased *standardGauge

	// Track memory growth rates
	lastHeapStats struct {
		timestamp time.Time
		heapInuse float64
	}
	heapGrowthRate *standardGauge // bytes/second

	// CPU metrics
	cpuUser   *standardGauge // User CPU time
	cpuSystem *standardGauge // System CPU time
	cpuIdle   *standardGauge // Idle CPU time

	// Disk metrics
	diskReads      *standardCounter // Number of read operations
	diskWrites     *standardCounter // Number of write operations
	diskReadBytes  *standardGauge   // Bytes read
	diskWriteBytes *standardGauge   // Bytes written

	lastCPUStats  *syscall.Rusage   // Last CPU stats for delta calculation
	lastDiskStats *syscall.Statfs_t // Last disk stats for delta calculation
	lastStatsTime time.Time

	// Thresholds for alerting
	thresholds Thresholds

	responseTimeTracker *responseTimeTracker
	recentRequests      *standardGauge // Requests in last minute
	requestsLastMinute  uint64         // For rate calculation
	lastMinuteCheck     time.Time
}

// StandardCollectorOption is a functional option for configuring a StandardCollector
type StandardCollectorOption func(*StandardCollector)

// WithServerName sets the server name for the collector
func WithServerName(name string) StandardCollectorOption {
	return func(c *StandardCollector) {
		c.serverName = name
	}
}

// WithThresholds sets the alert thresholds for the collector
func WithThresholds(thresholds Thresholds) StandardCollectorOption {
	return func(c *StandardCollector) {
		c.thresholds = thresholds
	}
}

// NewStandardCollector creates a new StandardCollector
func NewStandardCollector(opts ...StandardCollectorOption) *StandardCollector {
	c := &StandardCollector{
		serverName:          "HOP Server",
		startTime:           time.Now(),
		counters:            make(map[string]*standardCounter),
		gauges:              make(map[string]*standardGauge),
		histograms:          make(map[string]*standardHistogram),
		thresholds:          DefaultThresholds,
		lastStatsTime:       time.Now(),
		responseTimeTracker: newResponseTimeTracker(1000), // Keep last 1000 samples
		lastMinuteCheck:     time.Now(),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Initialize CPU metrics
	c.cpuUser = c.getOrCreateGauge("cpu_user_percent")
	c.cpuSystem = c.getOrCreateGauge("cpu_system_percent")
	c.cpuIdle = c.getOrCreateGauge("cpu_idle_percent")

	// Initialize disk metrics
	c.diskReads = c.getOrCreateCounter("disk_reads_total")
	c.diskWrites = c.getOrCreateCounter("disk_writes_total")
	c.diskReadBytes = c.getOrCreateGauge("disk_read_bytes")
	c.diskWriteBytes = c.getOrCreateGauge("disk_write_bytes")

	// Initialize common metrics
	c.httpRequests = c.getOrCreateCounter("http_requests_total")
	c.httpDurations = c.getOrCreateHistogram("http_request_duration_ms")
	c.httpServerErrors = c.getOrCreateCounter("http_errors_total")
	c.httpClientErrors = c.getOrCreateCounter("http_client_errors_total")

	c.goroutines = c.getOrCreateGauge("goroutines")
	//c.memAlloc = c.getOrCreateGauge("memory_alloc_bytes")
	//c.memTotal = c.getOrCreateGauge("memory_total_bytes")
	//c.memSys = c.getOrCreateGauge("memory_sys_bytes")
	c.gcPauses = c.getOrCreateHistogram("gc_pause_ms")

	c.heapInuse = c.getOrCreateGauge("memory_heap_inuse_bytes")
	c.heapSys = c.getOrCreateGauge("memory_heap_sys_bytes")
	c.heapIdle = c.getOrCreateGauge("memory_heap_idle_bytes")
	c.heapReleased = c.getOrCreateGauge("memory_heap_released_bytes")
	c.heapGrowthRate = c.getOrCreateGauge("memory_heap_growth_rate_bytes_per_sec")

	c.recentRequests = c.getOrCreateGauge("http_requests_last_minute")

	// Get initial stats
	c.lastCPUStats = &syscall.Rusage{}
	c.lastDiskStats = &syscall.Statfs_t{}
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, c.lastCPUStats)
	_ = syscall.Statfs(".", c.lastDiskStats)
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

	//c.memAlloc.Set(float64(ms.Alloc))
	//c.memTotal.Set(float64(ms.TotalAlloc))
	//c.memSys.Set(float64(ms.Sys))
	c.gcPauses.Observe(float64(ms.PauseNs[(ms.NumGC+255)%256]) / 1e6) // Convert to milliseconds

	// Update enhanced metrics
	c.heapInuse.Set(float64(ms.HeapInuse))
	c.heapSys.Set(float64(ms.HeapSys))
	c.heapIdle.Set(float64(ms.HeapIdle))
	c.heapReleased.Set(float64(ms.HeapReleased))

	// Calculate heap growth rate
	now := time.Now()
	if !c.lastHeapStats.timestamp.IsZero() {
		duration := now.Sub(c.lastHeapStats.timestamp).Seconds()
		if duration > 0 {
			growthRate := (float64(ms.HeapInuse) - c.lastHeapStats.heapInuse) / duration
			c.heapGrowthRate.Set(growthRate)
		}
	}

	// Update last stats
	c.lastHeapStats.timestamp = now
	c.lastHeapStats.heapInuse = float64(ms.HeapInuse)
}

// MemoryStatus represents the status of a specific memory metric
type MemoryStatus struct {
	Level     ThresholdLevel
	Reason    string
	Current   float64
	Threshold float64
	TrendInfo string // Additional information about trends
}

// checkMemoryThresholds evaluates memory usage against configured thresholds
func (c *StandardCollector) checkMemoryMetrics() map[string]MemoryStatus {
	status := make(map[string]MemoryStatus)
	status["memory_growth"] = MemoryStatus{}
	status["gc_pause"] = MemoryStatus{}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// Calculate memory growth rate (bytes per minute)
	memoryGrowthRate := c.heapGrowthRate.Value() * 60 // Convert per second to per minute
	growthPercent := (memoryGrowthRate / float64(ms.Sys)) * 100

	// Check memory growth
	memGrowthStatus := MemoryStatus{
		Level:     ThresholdOK,
		Current:   growthPercent,
		Threshold: c.thresholds.MemoryGrowthRatePercent,
		Reason:    fmt.Sprintf("%.1f%% per minute", growthPercent),
	}

	if growthPercent > c.thresholds.MemoryGrowthRatePercent {
		memGrowthStatus.Level = ThresholdWarning
		memGrowthStatus.Reason = fmt.Sprintf("Memory growing at %.1f%% per minute", growthPercent)
	}

	status["memory_growth"] = memGrowthStatus

	// Check GC metrics
	gcPauseMs := float64(ms.PauseNs[(ms.NumGC+255)%256]) / 1e6 // Convert ns to ms
	gcPauseStatus := MemoryStatus{
		Level:     ThresholdOK,
		Current:   gcPauseMs,
		Threshold: c.thresholds.GCPauseMs,
		Reason:    fmt.Sprintf("%.2fms", gcPauseMs),
	}

	if gcPauseMs > c.thresholds.GCPauseMs {
		gcPauseStatus.Level = ThresholdWarning
		gcPauseStatus.Reason = fmt.Sprintf("GC pause of %.2fms exceeds threshold", gcPauseMs)
	}

	status["gc_pause"] = gcPauseStatus

	// Calculate GC frequency (per minute)
	gcFrequency := float64(ms.NumGC) / time.Since(c.startTime).Minutes()
	gcFrequencyStatus := MemoryStatus{
		Level:     ThresholdOK,
		Current:   gcFrequency,
		Threshold: c.thresholds.MaxGCFrequency,
		Reason:    fmt.Sprintf("%.1f/minute", gcFrequency),
	}

	if gcFrequency > c.thresholds.MaxGCFrequency {
		gcFrequencyStatus.Level = ThresholdWarning
		gcFrequencyStatus.Reason = fmt.Sprintf("High GC frequency: %.1f/minute", gcFrequency)
	}

	status["gc_frequency"] = gcFrequencyStatus

	return status
}

// RecordGoroutineCount captures the number of goroutines
func (c *StandardCollector) RecordGoroutineCount() {
	c.goroutines.Set(float64(runtime.NumGoroutine()))
}

// RecordHTTPRequest records metrics about an HTTP request
func (c *StandardCollector) RecordHTTPRequest(method, path string, duration time.Duration, statusCode int) {
	c.httpRequests.Inc()
	c.httpDurations.Observe(float64(duration.Milliseconds()))
	c.responseTimeTracker.Record(float64(duration.Milliseconds()))

	// Update error count if status >= 400
	if statusCode >= 500 {
		c.httpServerErrors.Inc()
	} else if statusCode >= 400 {
		c.httpClientErrors.Inc()
	}

	// Update recent requests
	atomic.AddUint64(&c.requestsLastMinute, 1)

	// Update per-minute stats if needed
	now := time.Now()
	c.mu.Lock()
	if now.Sub(c.lastMinuteCheck) >= time.Minute {
		count := atomic.SwapUint64(&c.requestsLastMinute, 0)
		c.recentRequests.Set(float64(count))
		c.lastMinuteCheck = now
	}
	c.mu.Unlock()
}

// RecordCPUStats collects CPU usage statistics
func (c *StandardCollector) RecordCPUStats() {
	var currentStats syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &currentStats); err != nil {
		return
	}

	now := time.Now()
	duration := now.Sub(c.lastStatsTime).Seconds()

	if duration > 0 {
		// Calculate CPU usage percentages
		userTime := timeDiff(currentStats.Utime, c.lastCPUStats.Utime)
		systemTime := timeDiff(currentStats.Stime, c.lastCPUStats.Stime)

		userPercent := (userTime.Seconds() / duration) * 100
		systemPercent := (systemTime.Seconds() / duration) * 100
		idlePercent := 100 - (userPercent + systemPercent)

		c.cpuUser.Set(userPercent)
		c.cpuSystem.Set(systemPercent)
		c.cpuIdle.Set(idlePercent)
	}

	*c.lastCPUStats = currentStats
	c.lastStatsTime = now
}

// RecordDiskStats collects disk space usage statistics
func (c *StandardCollector) RecordDiskStats() {
	var currentStats syscall.Statfs_t
	if err := syscall.Statfs(".", &currentStats); err != nil {
		return
	}

	// Calculate space usage
	totalBytes := currentStats.Blocks * uint64(currentStats.Bsize)
	freeBytes := currentStats.Bfree * uint64(currentStats.Bsize)
	usedBytes := totalBytes - freeBytes

	// Update gauges for current disk space usage
	c.diskReadBytes.Set(float64(totalBytes)) // Total space
	c.diskWriteBytes.Set(float64(usedBytes)) // Used space

	// For the counters, we'll increment by the change in usage since last check
	if c.lastDiskStats != nil {
		lastTotal := c.lastDiskStats.Blocks * uint64(c.lastDiskStats.Bsize)
		lastFree := c.lastDiskStats.Bfree * uint64(c.lastDiskStats.Bsize)
		lastUsed := lastTotal - lastFree

		if usedBytes > lastUsed {
			c.diskWrites.Add(float64(usedBytes - lastUsed))
		}
		if totalBytes > lastTotal {
			c.diskReads.Add(float64(totalBytes - lastTotal))
		}
	}

	// Store current stats for next comparison
	c.lastDiskStats = &currentStats
}

// Helper function to calculate time difference
func timeDiff(a, b syscall.Timeval) time.Duration {
	sec := int64(a.Sec) - int64(b.Sec)
	usec := int64(a.Usec) - int64(b.Usec)
	return time.Duration(sec)*time.Second + time.Duration(usec)*time.Microsecond
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
