package metrics

import (
	"expvar"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"time"
)

// metricsHTML is a basic HTML template for displaying metrics
var metricsHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics - {{.ServerName}}</title>
    <meta charset='utf-8'>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.5; max-width: 1200px; margin: 0 auto; padding: 1rem; }
        h1 { color: #2d3748; border-bottom: 1px solid #e2e8f0; padding-bottom: 0.5rem; }
        h2 { color: #4a5568; margin-top: 2rem; }
        .metric-group { margin: 1rem 0; padding: 1rem; background: #f7fafc; border-radius: 0.5rem; }
        .metric { margin: 0.5rem 0; padding: 0.5rem 0; border-bottom: 1px solid #edf2f7; }
        .metric:last-child { border-bottom: none; }
        .metric-name { font-weight: bold; color: #4a5568; }
        .metric-value { font-family: monospace; color: #2b6cb0; }
        .metric-desc { display: block; margin-top: 0.25rem; color: #718096; font-size: 0.875rem; max-width: 500px; }
        .timestamp { color: #718096; font-size: 0.875rem; }
        .raw-link { float: right; color: #4a5568; text-decoration: none; }
        .raw-link:hover { text-decoration: underline; }
        .status-good { color: #48bb78; }
        .status-warning { color: #ecc94b; }
        .status-critical { color: #f56565; }        

        .refresh-control {
            position: fixed;
            top: 1rem;
            right: 1rem;
            background: white;
            padding: 0.5rem;
            border-radius: 0.5rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            display: flex;
            gap: 1rem;
            align-items: center;
        }
        .refresh-button {
            padding: 0.5rem 1rem;
            background: #4299e1;
            color: white;
            border: none;
            border-radius: 0.25rem;
            cursor: pointer;
        }
        .refresh-button:hover {
            background: #3182ce;
        }
        .chart {
            margin-top: 1rem;
            height: 50px;
            background: #f9fafb;
            border: 1px solid #e5e7eb;
            border-radius: 0.25rem;
        }

		@media (max-width: 480px) {
			.refresh-control {
				flex-direction: column;
			}
		}

        @media (max-width: 768px) {
            .refresh-control {
                position: static;
                margin-bottom: 1rem;
            }
        }

        .metric-value {
            font-family: monospace;
            padding: 0.2rem 0.5rem;
            border-radius: 0.25rem;
        }
        
		/* blue info level */
	    .level-0 .metric-value {
			color: #2b6cb0;
			background: #ebf8ff;
		}

        .level-1 .metric-value {
            color: #057a55;
            background: #def7ec;
        }
        
        .level-2 .metric-value {
            color: #c27803;
            background: #fdf6b2;
        }
        
        .level-3 .metric-value {
            color: #e02424;
            background: #fde8e8;
        }
        
        .threshold-info {
            font-size: 0.8rem;
            color: #6b7280;
            margin-left: 0.5rem;
        }
	</style>
    <script>
        let autoRefreshInterval = null;
        
        function refreshMetrics() {
            fetch(window.location.pathname + '?' + new URLSearchParams({
                format: 'html',
                _: Date.now() // Cache buster
            }))
            .then(response => response.text())
            .then(html => {
                const parser = new DOMParser();
                const newDoc = parser.parseFromString(html, 'text/html');
                
                // Update each metric group
                document.querySelectorAll('.metric-group').forEach(group => {
                    const newGroup = newDoc.querySelector(
                        '.metric-group:nth-of-type(' + 
                        Array.from(group.parentElement.children).indexOf(group) + 
                        ')'
                    );
                    if (newGroup) {
                        group.innerHTML = newGroup.innerHTML;
                    }
                });
                
                // Update timestamp
                const timestamp = document.querySelector('.timestamp');
                const newTimestamp = newDoc.querySelector('.timestamp');
                if (timestamp && newTimestamp) {
                    timestamp.textContent = newTimestamp.textContent;
                }
            })
            .catch(error => console.error('Error refreshing metrics:', error));
        }
        
        function toggleAutoRefresh() {
            const button = document.getElementById('autoRefreshButton');
            const interval = document.getElementById('refreshInterval');
            
            if (autoRefreshInterval) {
                clearInterval(autoRefreshInterval);
                autoRefreshInterval = null;
                button.textContent = 'Start Auto-refresh';
                interval.disabled = false;
            } else {
                const seconds = parseInt(interval.value);
                autoRefreshInterval = setInterval(refreshMetrics, seconds * 1000);
                button.textContent = 'Stop Auto-refresh';
                interval.disabled = true;
            }
        }
        
        // Stop auto-refresh when page is hidden
        document.addEventListener('visibilitychange', () => {
            if (document.hidden && autoRefreshInterval) {
                toggleAutoRefresh();
            }
        });
    </script>
</head>
<body>
    <div class="refresh-control">
        <select id="refreshInterval">
            <option value="5">5 seconds</option>
            <option value="10" selected>10 seconds</option>
            <option value="30">30 seconds</option>
            <option value="60">60 seconds</option>
        </select>
        <button id="autoRefreshButton" class="refresh-button" onclick="toggleAutoRefresh()">
            Start Auto-refresh
        </button>
		<a href="?format=json" class="raw-link">View Raw JSON</a>
    </div>

    <h1>System Metrics</h1>
    <div class="timestamp">Last Updated: {{.Timestamp}}</div>

<div class="metric-group">
        <h2>HTTP Metrics</h2>
        {{range .HTTPMetrics}}
        <div class="metric level-{{.Level}}">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
			{{if .Threshold}}<span class="threshold-info">Threshold: {{.Threshold}}</span>{{end}}
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>Memory Metrics</h2>
        {{range .MemoryMetrics}}
        <div class="metric level-{{.Level}}">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
			{{if .Threshold}}<span class="threshold-info">Threshold: {{.Threshold}}</span>{{end}}
			{{if .Reason}}<span class="threshold-info">Reason: {{.Reason}}</span>{{end}}
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>Runtime Metrics</h2>
        {{range .RuntimeMetrics}}
        <div class="metric level-{{.Level}}">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
			{{if .Threshold}}<span class="threshold-info">Threshold: {{.Threshold}}</span>{{end}}
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>CPU Metrics</h2>
        {{range .CPUMetrics}}
        <div class="metric level-{{.Level}}">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
			{{if .Threshold}}<span class="threshold-info">Threshold: {{.Threshold}}</span>{{end}}
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>Disk I/O Metrics</h2>
        {{range .DiskMetrics}}
        <div class="metric level-{{.Level}}">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
			{{if .Threshold}}<span class="threshold-info">Threshold: {{.Threshold}}</span>{{end}}
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    {{if .CustomMetrics}}
    <div class="metric-group">
        <h2>Custom Metrics</h2>
        {{range .CustomMetrics}}
        <div class="metric level-{{.Level}}">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
			{{if .Threshold}}<span class="threshold-info">Threshold: {{.Threshold}}</span>{{end}}
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>
    {{end}}

</body>
</html>
`

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
	tmpl := template.Must(template.New("metrics").Parse(metricsHTML))

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

		//// HTTP Metrics
		//reqCount := c.httpRequests.Value()
		//errCount := c.httpServerErrors.Value()
		//errRate := 0.0
		//if reqCount > 0 {
		//	errRate = (errCount / reqCount) * 100
		//}
		//avgDuration := 0.0
		//if c.httpDurations.Count() > 0 {
		//	avgDuration = c.httpDurations.Sum() / float64(c.httpDurations.Count())
		//}
		//
		//errorLevel := ThresholdOK
		//if errRate >= c.thresholds.ErrorRatePercent {
		//	errorLevel = ThresholdCritical
		//} else if errRate >= c.thresholds.ErrorRatePercent*0.5 {
		//	errorLevel = ThresholdWarning
		//}

		//data.HTTPMetrics = []metricData{
		//	{
		//		Name:        "Total Requests",
		//		Value:       formatCount(reqCount),
		//		Description: "Total number of HTTP requests processed since startup. Useful for understanding overall traffic patterns and calculating error rates. A sudden drop might indicate connectivity issues.",
		//		Level:       ThresholdInfo,
		//	},
		//	{
		//		Name:        "Error Count",
		//		Value:       formatCount(errCount),
		//		Description: fmt.Sprintf("Number of requests resulting in HTTP status codes >= 400 (%.1f%% error rate). Includes client errors (4xx) and server errors (5xx). Sustained high error rates might indicate application issues, bad client requests, or system problems.", errRate),
		//		Level:       errorLevel,
		//		Threshold:   fmt.Sprintf("%.1f%%", c.thresholds.ErrorRatePercent),
		//	},
		//	{
		//		Name:        "Average Response Time",
		//		Value:       formatDuration(avgDuration),
		//		Description: "Average time to process HTTP requests from receipt to response. Includes application processing time, database queries, and external service calls. Increasing response times might indicate performance bottlenecks or resource constraints.",
		//		Level:       ThresholdInfo,
		//	},
		//	{
		//		Name:        "Requests/Second",
		//		Value:       fmt.Sprintf("%.2f", float64(reqCount)/time.Since(c.startTime).Seconds()),
		//		Description: "Average request throughput since server start. Useful for capacity planning and load balancing. Compare with historical patterns to identify unusual traffic patterns.",
		//		Level:       ThresholdInfo,
		//	},
		//}
		data.HTTPMetrics = c.formatHTTPMetrics()

		// Memory Metrics
		//memUsed := c.memAlloc.Value()
		//memTotal := c.memSys.Value()
		//memPercent := (memUsed / memTotal) * 100
		//memLevel := ThresholdOK
		//if memPercent >= c.thresholds.MemoryPercent {
		//	memLevel = ThresholdCritical
		//} else if memPercent >= c.thresholds.MemoryPercent*0.8 {
		//	memLevel = ThresholdWarning
		//}
		//
		//data.MemoryMetrics = []metricData{
		//	{
		//		Name:        "Allocated Memory",
		//		Value:       formatBytes(c.memAlloc.Value()),
		//		Description: "Current heap memory actively allocated by the application that hasn't been freed. Lower than System Memory because it only includes live objects in the heap.",
		//		Level:       memLevel,
		//		Threshold:   fmt.Sprintf("%.1f%%", c.thresholds.MemoryPercent),
		//	},
		//	{
		//		Name:        "Total Allocated",
		//		Value:       formatBytes(c.memTotal.Value()),
		//		Description: "Cumulative memory allocated since start. This includes both currently allocated memory and memory that has been freed - useful for understanding memory allocation patterns over time. Unlike Allocated Memory, this never decreases. Useful for understanding the total memory throughput of your application over time.",
		//		Level:       ThresholdInfo,
		//	},
		//	{
		//		Name:        "System Memory",
		//		Value:       formatBytes(c.memSys.Value()),
		//		Description: "Total memory obtained from the operating system by the Go runtime, including heap, stack, memory for goroutines, and other runtime structures. This is the actual memory footprint of your application.",
		//		Level:       ThresholdInfo,
		//	},
		//	{
		//		Name:        "Heap Objects",
		//		Value:       formatCount(float64(ms.HeapObjects)),
		//		Description: "Number of objects currently allocated on the heap. A steady increase might indicate a memory leak, while frequent large variations might suggest heavy garbage collection activity. Correlates with Allocated Memory but counts objects rather than bytes. High numbers might indicate many small allocations.",
		//		Level:       ThresholdInfo,
		//	},
		//	{
		//		Name:        "GC Cycles",
		//		Value:       formatCount(float64(ms.NumGC)),
		//		Description: "Number of completed garbage collection cycles since startup. Frequent GC cycles might indicate memory pressure or inefficient memory usage patterns. In healthy applications, GC frequency should correlate with allocation rates and available memory. Sustained high frequency might indicate memory pressure.",
		//		Level:       ThresholdInfo,
		//	},
		//}

		data.MemoryMetrics = c.formatMemoryMetrics()

		// Runtime Metrics
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

		data.RuntimeMetrics = []metricData{
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

		cpuUsed := 100 - c.cpuIdle.Value()
		cpuLevel := ThresholdOK
		if cpuUsed >= c.thresholds.CPUPercent {
			cpuLevel = ThresholdCritical
		} else if cpuUsed >= c.thresholds.CPUPercent*0.8 {
			cpuLevel = ThresholdWarning
		}

		data.CPUMetrics = []metricData{
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

		diskUsed := c.diskWriteBytes.Value()
		diskTotal := c.diskReadBytes.Value()
		diskPercent := (diskUsed / diskTotal) * 100
		diskLevel := ThresholdOK
		if diskPercent >= c.thresholds.DiskPercent {
			diskLevel = ThresholdCritical
		} else if diskPercent >= c.thresholds.DiskPercent*0.8 {
			diskLevel = ThresholdWarning
		}

		data.DiskMetrics = []metricData{
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
