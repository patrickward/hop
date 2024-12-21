# Pulse Package

The `pulse` package provides a quick and dirty way to collect and monitor application metrics in Go applications using the hop framework. It includes HTTP middleware for request metrics, memory statistics tracking, and a built-in metrics visualization dashboard. It's not meant to be a full-fledged monitoring solution but rather a simple way to get started with metrics collection. For more advanced use cases, consider using a dedicated monitoring tool like Prometheus or Grafana. 

## Features

- Real-time metrics collection and monitoring
- Built-in metrics dashboard with auto-refresh capabilities
- HTTP middleware for request metrics
- Memory, CPU, and disk usage tracking
- Configurable thresholds for alerts
- Optional pprof endpoints integration
- JSON and HTML visualization formats

## Quick Start

```go
// Create a new metrics collector
collector := pulse.NewStandardCollector(
    pulse.WithServerName("MyApp"),
    pulse.WithThresholds(pulse.Thresholds{
        CPUPercent: 80.0,
        MemoryPercent: 85.0,
    }),
)

// Create and configure the pulse module
pulseMod := pulse.NewModule(collector, &pulse.Config{
    EnablePprof: !app.Config().IsProduction(),
    PulsePath: "/pulse",  // Default path
    CollectionInterval: 15 * time.Second,  // Default interval
})

// Register with your application
app.RegisterModule(pulseMod)

// Add the middleware to collect HTTP pulse
app.Router().Use(pulseMod.Middleware())
```

## Default Thresholds

The package comes with pre-configured default thresholds that can be customized:

```go
var DefaultThresholds = pulse.Thresholds{
    CPUPercent:              75.0,  // Warning at 75% CPU usage
    ClientErrorRatePercent:  40.0,  // Higher threshold for 4xx errors
    DiskPercent:             85.0,  // Warning at 85% disk usage
    GCPauseMs:              100.0,  // 100ms pause time warning
    GoroutineCount:         1000,   // Warning at 1000 goroutines
    MaxGCFrequency:         100.0,  // Warning at >100 GCs/minute
    MemoryGrowthRatePercent: 20.0,  // Warning at 20% growth/minute
    MemoryPercent:           80.0,  // Warning at 80% memory usage
    ServerErrorRatePercent:   1.0,  // Very low tolerance for 5xx errors
}
```

### Customizing Thresholds

You can customize thresholds when creating the collector:

```go
collector := pulse.NewStandardCollector(
    pulse.WithThresholds(pulse.Thresholds{
        CPUPercent: 90.0,                // More lenient CPU threshold
        MemoryPercent: 90.0,             // More lenient memory threshold
        ServerErrorRatePercent: 0.5,     // Stricter error threshold
        GoroutineCount: 2000,            // Allow more goroutines
    }),
)
```

## Pulse Dashboard

The pulse dashboard is available at `/pulse` by default (configurable via `PulsePath`). It provides:

### HTTP Metrics
- Total request count
- Recent and overall request rates
- Client (4xx) and Server (5xx) error rates
- Response time percentiles (P95, P99)
- Average response time

### Memory Metrics
- Application memory usage
- Memory growth rate
- Garbage collection statistics
- Heap usage and utilization

### Runtime Metrics
- Active goroutine count
- GC pause times
- CPU thread count
- Application uptime

### CPU Metrics
- User CPU time
- System CPU time
- Idle CPU time

### Disk I/O Metrics
- Total disk space
- Used space
- Space growth metrics

### Features of the Dashboard
- Auto-refresh capabilities (configurable intervals)
- Color-coded thresholds for quick status checks
- Detailed descriptions for each metric
- Raw JSON data access
- Mobile-responsive design

## Debug/Development Features

When `EnablePprof` is set to true (recommended for non-production environments), additional debug endpoints are available:

- `/debug/pprof/` - Index of pprof endpoints
- `/debug/pprof/cmdline` - Command line arguments
- `/debug/pprof/profile` - CPU profile
- `/debug/pprof/symbol` - Symbol lookup
- `/debug/pprof/trace` - Execution trace

## Metrics Levels

Metrics are displayed with different levels based on their thresholds:

- **Info** (Blue) - General information
- **OK** (Green) - Within normal range
- **Warning** (Yellow) - Approaching threshold
- **Critical** (Red) - Exceeded threshold

## Best Practices

1. **Production Setup**
   ```go
   pulseMod := pulse.NewModule(collector, &pulse.Config{
       EnablePprof: false,  // Disable pprof in production
       CollectionInterval: 30 * time.Second,  // Adjust based on needs
   })
   ```

2. **Development Setup**
   ```go
   pulseMod := pulse.NewModule(collector, &pulse.Config{
       EnablePprof: true,  // Enable debugging tools
       CollectionInterval: 5 * time.Second,  // More frequent updates
   })
   ```

3. **Custom Thresholds**: Adjust thresholds based on your application's characteristics and requirements.

4. **Security**: Consider adding authentication middleware for the pulse endpoint in production environments.

## Implementation Details

The package uses:
- `expvar` for metrics storage
- `runtime` package for memory statistics
- `syscall` for CPU and disk metrics
- Standard library's HTTP server for the dashboard

## Notes

- The pulse dashboard is designed to be lightweight and doesn't require external dependencies
- All metrics are collected in-memory
- The dashboard uses vanilla JavaScript for auto-refresh functionality
- Metric collection has minimal performance impact
- The middleware automatically tracks HTTP request metrics
