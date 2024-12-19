// Package metrics provides standardized metrics collection for hop applications
package metrics

import (
	"net/http"
	"time"
)

// Collector defines the interface for metrics collection
type Collector interface {
	// Core metric types

	// Counter is for cumulative metrics that only increase
	Counter(name string) Counter
	// Gauge is for metrics that can go up and down
	Gauge(name string) Gauge
	// Histogram tracks the distribution of a metric
	Histogram(name string) Histogram

	// System metrics

	// RecordMemStats records memory statistics
	RecordMemStats()
	// RecordGoroutineCount records the number of goroutines
	RecordGoroutineCount()

	// HTTP metrics

	// RecordHTTPRequest records an HTTP request
	RecordHTTPRequest(method, path string, duration time.Duration, statusCode int)

	// GetHandler returns an http.Handler for the metrics endpoint

	// Handler returns an http.Handler for the metrics endpoint
	Handler() http.Handler

	// HandlerJSON returns an http.Handler for the raw JSON metrics endpoint
	HandlerJSON() http.Handler
}

// Counter is for cumulative metrics that only increase
type Counter interface {
	Inc()
	Add(delta float64)
	Value() float64
}

// Gauge is for metrics that can go up and down
type Gauge interface {
	Set(value float64)
	Add(delta float64)
	Sub(delta float64)
	Value() float64
}

// Histogram tracks the distribution of a metric
type Histogram interface {
	Observe(value float64)
	Count() uint64
	Sum() float64
}

// Labels represents a set of metric labels/tags
type Labels map[string]string
