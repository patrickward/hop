package metrics

import (
	"embed"
	"expvar"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

type metricData struct {
	Name        string
	Value       string
	Description string
	Level       ThresholdLevel
	Threshold   string
	Reason      string
}

type presentedMetrics struct {
	ServerName     string
	Timestamp      string
	HTTPMetrics    []metricData
	MemoryMetrics  []metricData
	RuntimeMetrics []metricData
	CustomMetrics  []metricData
	CPUMetrics     []metricData
	DiskMetrics    []metricData
}

// Handler returns an http.Handler for the metrics endpoint as an HTML page
func (c *StandardCollector) Handler() http.Handler {
	tmpl := template.Must(template.New("metrics").ParseFS(templateFS, "templates/metrics.html"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Force collection of current metrics
		c.RecordMemStats()
		c.RecordGoroutineCount()
		c.RecordCPUStats()
		c.RecordDiskStats()

		if r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			expvar.Handler().ServeHTTP(w, r)
			return
		}

		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)

		data := presentedMetrics{
			ServerName: c.serverName,
			Timestamp:  time.Now().Format("2006-01-02 15:04:05 MST"),
		}

		data.HTTPMetrics = c.formatHTTPMetrics()
		data.MemoryMetrics = c.formatMemoryMetrics()
		data.RuntimeMetrics = c.formatRuntimeMetrics()
		data.CPUMetrics = c.formatCPUMetrics()
		data.DiskMetrics = c.formatDiskMetrics()

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Error rendering metrics page: "+err.Error(), http.StatusInternalServerError)
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

//func (c *StandardCollector) formatCustomMetrics() []metricData {
//	var metrics []metricData
//
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//
//	// Add counters
//	for name, counter := range c.counters {
//		if !isSystemMetric(name) {
//			metrics = append(metrics, metricData{
//				Name:  formatMetricName(name),
//				Value: formatCount(counter.Value()),
//			})
//		}
//	}
//
//	// Add gauges
//	for name, gauge := range c.gauges {
//		if !isSystemMetric(name) {
//			metrics = append(metrics, metricData{
//				Name:  formatMetricName(name),
//				Value: fmt.Sprintf("%.2f", gauge.Value()),
//			})
//		}
//	}
//
//	// Sort metrics by name for consistent display
//	sort.Slice(metrics, func(i, j int) bool {
//		return metrics[i].Name < metrics[j].Name
//	})
//
//	return metrics
//}

//func formatMetricName(name string) string {
//	// Convert snake_case to Title Case
//	parts := strings.Split(name, "_")
//	for i, part := range parts {
//		parts[i] = strings.Title(part)
//	}
//	return strings.Join(parts, " ")
//}

//func isSystemMetric(name string) bool {
//	systemPrefixes := []string{
//		"http_",
//		"memory_",
//		"goroutines",
//		"gc_",
//	}
//
//	for _, prefix := range systemPrefixes {
//		if strings.HasPrefix(name, prefix) {
//			return true
//		}
//	}
//	return false
//}

func calculateErrorLevel(rate, threshold float64) ThresholdLevel {
	if rate >= threshold {
		return ThresholdCritical
	} else if rate >= threshold*0.5 {
		return ThresholdWarning
	}
	return ThresholdOK
}

func (c *StandardCollector) formatHTTPMetrics() []metricData {
	reqCount := c.httpRequests.Value()
	clientErrors := c.httpClientErrors.Value()
	serverErrors := c.httpServerErrors.Value()
	clientErrorRate := 0.0
	serverErrorRate := 0.0
	if reqCount > 0 {
		clientErrorRate = (clientErrors / reqCount) * 100
		serverErrorRate = (serverErrors / reqCount) * 100
	}

	// Get response time metrics
	p95 := c.responseTimeTracker.GetPercentile(95)
	p99 := c.responseTimeTracker.GetPercentile(99)
	avg := c.responseTimeTracker.GetAverage()

	// Calculate request rates
	recentRate := c.recentRequests.Value()
	overallRate := float64(reqCount) / time.Since(c.startTime).Seconds()

	return []metricData{
		{
			Name:        "Total Requests",
			Value:       formatCount(reqCount),
			Description: "Total number of HTTP requests processed since startup.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Recent Request Rate",
			Value:       fmt.Sprintf("%.1f/min", recentRate),
			Description: "Number of requests in the last minute. Compare with overall rate to identify traffic spikes.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Overall Request Rate",
			Value:       fmt.Sprintf("%.2f/sec", overallRate),
			Description: "Average requests per second since startup.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Client Errors (4xx)",
			Value:       fmt.Sprintf("%.1f%% (%s errors)", clientErrorRate, formatCount(clientErrors)),
			Description: "Percentage of requests resulting in 4xx status codes. Usually indicates client-side issues like validation errors or missing resources.",
			Level:       calculateErrorLevel(clientErrorRate, c.thresholds.ClientErrorRatePercent),
			Threshold:   fmt.Sprintf("%.1f%%", c.thresholds.ClientErrorRatePercent),
		},
		{
			Name:        "Server Errors (5xx)",
			Value:       fmt.Sprintf("%.1f%% (%s errors)", serverErrorRate, formatCount(serverErrors)),
			Description: "Percentage of requests resulting in 5xx status codes. Indicates server-side problems that need investigation.",
			Level:       calculateErrorLevel(serverErrorRate, c.thresholds.ServerErrorRatePercent),
			Threshold:   fmt.Sprintf("%.1f%%", c.thresholds.ServerErrorRatePercent),
		},
		{
			Name:        "Response Time (P95)",
			Value:       fmt.Sprintf("%.2f ms", p95),
			Description: "95% of requests complete within this time. A better indicator of user experience than average.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Response Time (P99)",
			Value:       fmt.Sprintf("%.2f ms", p99),
			Description: "99% of requests complete within this time. Useful for identifying worst-case response times.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Average Response Time",
			Value:       fmt.Sprintf("%.2f ms", avg),
			Description: "Mean response time across all requests. May be skewed by outliers.",
			Level:       ThresholdInfo,
		},
	}
}

func (c *StandardCollector) formatMemoryMetrics() []metricData {
	status := c.checkMemoryMetrics()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// For trend calculation
	totalAppMemory := float64(ms.Sys)
	heapInUse := float64(ms.HeapInuse)

	return []metricData{
		{
			Name:        "Application Memory",
			Value:       fmt.Sprintf("%.1f MiB", totalAppMemory/(1<<20)),
			Description: "Total memory obtained from the OS for this Go application. This represents the application's total memory footprint.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Memory Growth Rate",
			Value:       fmt.Sprintf("%.1f MiB/min", (c.heapGrowthRate.Value()*60)/(1<<20)),
			Description: "Rate of change in total memory usage. Sustained increases may indicate memory leaks.",
			Level:       status["memory_growth"].Level,
			Threshold:   fmt.Sprintf("%.1f%%/min", c.thresholds.MemoryGrowthRatePercent),
			Reason:      status["memory_growth"].Reason,
		},
		{
			Name:        "GC Pause Time",
			Value:       status["gc_pause"].Reason,
			Description: "Time the application pauses for garbage collection. Long pauses can affect responsiveness.",
			Level:       status["gc_pause"].Level,
			Threshold:   fmt.Sprintf("%.1fms", c.thresholds.GCPauseMs),
			Reason:      status["gc_pause"].Reason,
		},
		{
			Name:        "GC Frequency",
			Value:       status["gc_frequency"].Reason,
			Description: "How often garbage collection runs. High frequency might indicate memory pressure.",
			Level:       status["gc_frequency"].Level,
			Threshold:   fmt.Sprintf("%.0f/min", c.thresholds.MaxGCFrequency),
			Reason:      status["gc_frequency"].Reason,
		},
		{
			Name:        "Heap Usage",
			Value:       fmt.Sprintf("%.1f MiB (%.1f%% utilized)", heapInUse/(1<<20), (heapInUse/float64(ms.HeapSys))*100),
			Description: "Current heap memory usage and utilization. Go's GC typically maintains high utilization (90-100%) for efficiency.",
			Level:       ThresholdInfo,
		},
	}
}

func (c *StandardCollector) formatRuntimeMetrics() []metricData {
	gcAvg := 0.0
	if c.gcPauses.Count() > 0 {
		gcAvg = c.gcPauses.Sum() / float64(c.gcPauses.Count())
	}

	goroutines := int(c.goroutines.Value())
	goroutineLevel := ThresholdOK
	if goroutines >= c.thresholds.GoroutineCount {
		goroutineLevel = ThresholdCritical
	} else if goroutines >= int(float64(c.thresholds.GoroutineCount)*0.8) {
		goroutineLevel = ThresholdWarning
	}

	return []metricData{
		{
			Name:        "Goroutines",
			Value:       formatCount(c.goroutines.Value()),
			Description: "Current number of goroutines in the application. A very high or constantly increasing number might indicate goroutine leaks or inefficient concurrency patterns.",
			Level:       goroutineLevel,
			Threshold:   formatCount(float64(c.thresholds.GoroutineCount)),
		},
		{
			Name:        "Average GC Pause",
			Value:       formatDuration(gcAvg),
			Description: "Average time the application paused for garbage collection. Long pauses can affect application responsiveness and latency. Lower is better.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "CPU Threads",
			Value:       formatCount(float64(runtime.NumCPU())),
			Description: "Number of CPU threads available to Go runtime. This is typically the number of logical CPU cores on the system, which affects parallel processing capability.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Uptime",
			Value:       formatDuration(float64(time.Since(c.startTime).Milliseconds())),
			Description: "Time elapsed since the application started. Useful for monitoring application restarts and calculating average metrics over time.",
			Level:       ThresholdInfo,
		},
	}
}

func (c *StandardCollector) formatCPUMetrics() []metricData {
	cpuUsed := 100 - c.cpuIdle.Value()
	cpuLevel := ThresholdOK
	if cpuUsed >= c.thresholds.CPUPercent {
		cpuLevel = ThresholdCritical
	} else if cpuUsed >= c.thresholds.CPUPercent*0.8 {
		cpuLevel = ThresholdWarning
	}

	return []metricData{
		{
			Name:        "User CPU",
			Value:       fmt.Sprintf("%.1f%%", c.cpuUser.Value()),
			Description: "Percentage of CPU time spent executing application code (user space). High values indicate compute-intensive application workload.",
			Level:       cpuLevel,
			Threshold:   fmt.Sprintf("%.1f%%", c.thresholds.CPUPercent),
		},
		{
			Name:        "System CPU",
			Value:       fmt.Sprintf("%.1f%%", c.cpuSystem.Value()),
			Description: "Percentage of CPU time spent executing kernel code (system space). High values might indicate heavy I/O, system calls, or context switching.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Idle CPU",
			Value:       fmt.Sprintf("%.1f%%", c.cpuIdle.Value()),
			Description: "Percentage of CPU time where the processor was idle. Low values indicate high CPU utilization which might affect application performance.",
			Level:       ThresholdInfo,
		},
	}
}

func (c *StandardCollector) formatDiskMetrics() []metricData {
	diskUsed := c.diskWriteBytes.Value()
	diskTotal := c.diskReadBytes.Value()
	diskPercent := (diskUsed / diskTotal) * 100
	diskLevel := ThresholdOK
	if diskPercent >= c.thresholds.DiskPercent {
		diskLevel = ThresholdCritical
	} else if diskPercent >= c.thresholds.DiskPercent*0.8 {
		diskLevel = ThresholdWarning
	}

	return []metricData{
		{
			Name:        "Total Space",
			Value:       formatBytes(c.diskReadBytes.Value()),
			Description: "Total disk space available to the application's filesystem. This represents the total capacity of the volume where the application is running.",
			Level:       diskLevel,
			Threshold:   fmt.Sprintf("%.1f%%", c.thresholds.DiskPercent),
		},
		{
			Name:        "Used Space",
			Value:       formatBytes(c.diskWriteBytes.Value()),
			Description: "Amount of disk space currently in use. Includes application files, logs, and any data written by the application. High utilization might impact performance or cause write failures.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Space Added",
			Value:       formatBytes(c.diskWrites.Value()),
			Description: "Cumulative increase in disk space usage. Useful for tracking disk growth over time and predicting future capacity requirements.",
			Level:       ThresholdInfo,
		},
		{
			Name:        "Total Growth",
			Value:       formatBytes(c.diskReads.Value()),
			Description: "Cumulative growth in total disk space. Includes both space used and space freed. Useful for understanding disk usage patterns over time.",
			Level:       ThresholdInfo,
		},
	}
}
