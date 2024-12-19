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
        .metric-desc { display: block; margin-top: 0.25rem; color: #718096; font-size: 0.875rem; }
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
        @media (max-width: 768px) {
            .refresh-control {
                position: static;
                margin-bottom: 1rem;
            }
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
    </div>

	<a href="?format=json" class="raw-link">View Raw JSON</a>
    <h1>System Metrics</h1>
    <div class="timestamp">Last Updated: {{.Timestamp}}</div>

<div class="metric-group">
        <h2>HTTP Metrics</h2>
        {{range .HTTPMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>Memory Metrics</h2>
        {{range .MemoryMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>Runtime Metrics</h2>
        {{range .RuntimeMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>CPU Metrics</h2>
        {{range .CPUMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
            <span class="metric-desc">{{.Description}}</span>
        </div>
        {{end}}
    </div>

    <div class="metric-group">
        <h2>Disk I/O Metrics</h2>
        {{range .DiskMetrics}}
        <div class="metric">
            <span class="metric-name">{{.Name}}:</span>
            <span class="metric-value">{{.Value}}</span>
            <span class="metric-desc">{{.Description}}</span>
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

		// HTTP Metrics
		reqCount := c.httpRequests.Value()
		errCount := c.httpErrors.Value()
		errRate := 0.0
		if reqCount > 0 {
			errRate = (errCount / reqCount) * 100
		}
		avgDuration := 0.0
		if c.httpDurations.Count() > 0 {
			avgDuration = c.httpDurations.Sum() / float64(c.httpDurations.Count())
		}

		data.HTTPMetrics = []metricData{
			{
				Name:        "Total Requests",
				Value:       formatCount(reqCount),
				Description: "Total number of HTTP requests processed",
			},
			{
				Name:        "Error Count",
				Value:       formatCount(errCount),
				Description: fmt.Sprintf("Number of requests resulting in errors (%.1f%% error rate)", errRate),
			},
			{
				Name:        "Average Response Time",
				Value:       formatDuration(avgDuration),
				Description: "Average time to process HTTP requests",
			},
			{
				Name:        "Requests/Second",
				Value:       fmt.Sprintf("%.2f", float64(reqCount)/time.Since(c.startTime).Seconds()),
				Description: "Average number of requests per second since server start",
			},
		}

		// Memory Metrics
		data.MemoryMetrics = []metricData{
			{
				Name:        "Allocated Memory",
				Value:       formatBytes(c.memAlloc.Value()),
				Description: "Currently allocated memory in use",
			},
			{
				Name:        "Total Allocated",
				Value:       formatBytes(c.memTotal.Value()),
				Description: "Cumulative memory allocated since server start",
			},
			{
				Name:        "System Memory",
				Value:       formatBytes(c.memSys.Value()),
				Description: "Total memory obtained from the OS",
			},
			{
				Name:        "Heap Objects",
				Value:       formatCount(float64(ms.HeapObjects)),
				Description: "Number of allocated heap objects",
			},
			{
				Name:        "GC Cycles",
				Value:       formatCount(float64(ms.NumGC)),
				Description: "Number of completed garbage collection cycles",
			},
		}

		// Runtime Metrics
		gcAvg := 0.0
		if c.gcPauses.Count() > 0 {
			gcAvg = c.gcPauses.Sum() / float64(c.gcPauses.Count())
		}

		data.RuntimeMetrics = []metricData{
			{
				Name:        "Goroutines",
				Value:       formatCount(c.goroutines.Value()),
				Description: "Current number of running goroutines",
			},
			{
				Name:        "Average GC Pause",
				Value:       formatDuration(gcAvg),
				Description: "Average garbage collection pause duration",
			},
			{
				Name:        "CPU Threads",
				Value:       formatCount(float64(runtime.NumCPU())),
				Description: "Number of CPU threads available",
			},
			{
				Name:        "Uptime",
				Value:       formatDuration(float64(time.Since(c.startTime).Milliseconds())),
				Description: "Time since server start",
			},
		}

		data.CPUMetrics = []metricData{
			{
				Name:        "User CPU",
				Value:       fmt.Sprintf("%.1f%%", c.cpuUser.Value()),
				Description: "Percentage of CPU time spent in user space",
			},
			{
				Name:        "System CPU",
				Value:       fmt.Sprintf("%.1f%%", c.cpuSystem.Value()),
				Description: "Percentage of CPU time spent in kernel space",
			},
			{
				Name:        "Idle CPU",
				Value:       fmt.Sprintf("%.1f%%", c.cpuIdle.Value()),
				Description: "Percentage of CPU time idle",
			},
		}

		data.DiskMetrics = []metricData{
			{
				Name:        "Total Space",
				Value:       formatBytes(c.diskReadBytes.Value()),
				Description: "Total disk space available",
			},
			{
				Name:        "Used Space",
				Value:       formatBytes(c.diskWriteBytes.Value()),
				Description: "Total disk space currently in use",
			},
			{
				Name:        "Space Added",
				Value:       formatBytes(c.diskWrites.Value()),
				Description: "Cumulative increase in disk space usage",
			},
			{
				Name:        "Total Growth",
				Value:       formatBytes(c.diskReads.Value()),
				Description: "Cumulative growth in total disk space",
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
