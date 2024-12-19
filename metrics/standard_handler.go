package metrics

import (
	"expvar"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"
)

// metricsHTML is a basic HTML template for displaying metrics
var metricsHTML = `
<!DOCTYPE html>
<html>
<head>
	<meta charset='utf-8'>
	<meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Metrics - {{.ServerName}}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.5; max-width: 1200px; margin: 0 auto; padding: 1rem; }
        h1 { color: #2d3748; border-bottom: 1px solid #e2e8f0; padding-bottom: 0.5rem; }
        h2 { color: #4a5568; margin-top: 2rem; }
        .metric-group { margin: 1rem 0; padding: 1rem; background: #f7fafc; border-radius: 0.5rem; }
        .metric { margin: 0.5rem 0; }
        .metric-name { font-weight: bold; color: #4a5568; }
        .metric-value { font-family: monospace; color: #2b6cb0; }
        .timestamp { color: #718096; font-size: 0.875rem; }
        .raw-link { float: right; color: #4a5568; text-decoration: none; }
        .raw-link:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <a href="?format=json" class="raw-link">View Raw JSON</a>
    <h1>System Metrics</h1>
    <div class="timestamp">Last Updated: {{.Timestamp}}</div>

    <div class="metric-group">
        <h2>HTTP Metrics</h2>
        {{range .HTTPMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>Memory Metrics</h2>
        {{range .MemoryMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>Runtime Metrics</h2>
        {{range .RuntimeMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
        </div>
        {{end}}
    </div>

    {{if .CustomMetrics}}
    <div class="metric-group">
        <h2>Custom Metrics</h2>
        {{range .CustomMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>
`

type metricData struct {
	Name  string
	Value string
}

type presentedMetrics struct {
	ServerName     string
	Timestamp      string
	HTTPMetrics    []metricData
	MemoryMetrics  []metricData
	RuntimeMetrics []metricData
	CustomMetrics  []metricData
}

// HandlerJSON returns an http.Handler for the raw JSON metrics endpoint
func (c *StandardCollector) HandlerJSON() http.Handler {
	return expvar.Handler()
}

// Handler returns an http.Handler for the metrics endpoint as an HTML page
func (c *StandardCollector) Handler() http.Handler {
	tmpl := template.Must(template.New("metrics").Parse(metricsHTML))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Force collection of current metrics
		c.RecordMemStats()
		c.RecordGoroutineCount()

		// Check if raw JSON format is requested
		if r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			expvar.Handler().ServeHTTP(w, r)
			return
		}

		// Prepare metrics data
		data := presentedMetrics{
			ServerName: c.serverName,
			Timestamp:  time.Now().Format("2006-01-02 15:04:05 MST"),
		}

		// HTTP Metrics
		reqCount := c.httpRequests.Value()
		errCount := c.httpErrors.Value()
		avgDuration := 0.0
		if c.httpDurations.Count() > 0 {
			avgDuration = c.httpDurations.Sum() / float64(c.httpDurations.Count())
		}

		data.HTTPMetrics = []metricData{
			{"Total Requests", formatCount(reqCount)},
			{"Error Count", formatCount(errCount)},
			{"Average Response Time", formatDuration(avgDuration)},
		}

		// Memory Metrics
		data.MemoryMetrics = []metricData{
			{"Allocated Memory", formatBytes(c.memAlloc.Value())},
			{"Total Allocated", formatBytes(c.memTotal.Value())},
			{"System Memory", formatBytes(c.memSys.Value())},
		}

		// Runtime Metrics
		gcAvg := 0.0
		if c.gcPauses.Count() > 0 {
			gcAvg = c.gcPauses.Sum() / float64(c.gcPauses.Count())
		}

		data.RuntimeMetrics = []metricData{
			{"Goroutines", formatCount(c.goroutines.Value())},
			{"Average GC Pause", formatDuration(gcAvg)},
		}

		// Add any custom metrics
		data.CustomMetrics = c.formatCustomMetrics()

		// Serve HTML
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Error rendering metrics page", http.StatusInternalServerError)
			return
		}
	})
}

// Helper functions for formatting values
func formatBytes(bytes float64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%.1f B", bytes)
	}
	div, exp := unit, 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", bytes/float64(div), "KMGTPE"[exp])
}

func formatCount(value float64) string {
	if value < 1000 {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.1fk", value/1000)
}

func formatDuration(ms float64) string {
	if ms < 1 {
		return fmt.Sprintf("%.2f µs", ms*1000)
	}
	if ms < 1000 {
		return fmt.Sprintf("%.2f ms", ms)
	}
	return fmt.Sprintf("%.2f s", ms/1000)
}

func (c *StandardCollector) formatCustomMetrics() []metricData {
	var metrics []metricData

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Add counters
	for name, counter := range c.counters {
		if !isSystemMetric(name) {
			metrics = append(metrics, metricData{
				Name:  formatMetricName(name),
				Value: formatCount(counter.Value()),
			})
		}
	}

	// Add gauges
	for name, gauge := range c.gauges {
		if !isSystemMetric(name) {
			metrics = append(metrics, metricData{
				Name:  formatMetricName(name),
				Value: fmt.Sprintf("%.2f", gauge.Value()),
			})
		}
	}

	// Sort metrics by name for consistent display
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	return metrics
}

func formatMetricName(name string) string {
	// Convert snake_case to Title Case
	parts := strings.Split(name, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, " ")
}

func isSystemMetric(name string) bool {
	systemPrefixes := []string{
		"http_",
		"memory_",
		"goroutines",
		"gc_",
	}

	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
