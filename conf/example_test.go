package conf_test

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/patrickward/hop/conf"
)

func Example() {
	// STEP1: Define a custom configuration struct that embeds BaseConfig
	type AppConfig struct {
		conf.BaseConfig // Inherit base configuration
		Redis           struct {
			Host    string        `json:"host" env:"REDIS_HOST" default:"localhost"`
			Port    int           `json:"port" env:"REDIS_PORT" default:"6379"`
			Timeout conf.Duration `json:"timeout" env:"REDIS_TIMEOUT" default:"5s"`
		} `json:"redis"`
		API struct {
			Endpoint string        `json:"endpoint" env:"API_ENDPOINT" default:"http://api.local"`
			Timeout  conf.Duration `json:"timeout" env:"API_TIMEOUT" default:"30s"`
		} `json:"api"`
	}

	// Create a temporary config file (for example purposes only)
	configJSON := `{
        "environment": "production",
        "debug": false,
        "redis": {
            "host": "redis.prod.example.com",
            "timeout": "10s"
        },
        "api": {
            "endpoint": "https://api.prod.example.com"
        }
    }`
	tmpfile, err := os.CreateTemp("", "config.*.json")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configJSON)); err != nil {
		fmt.Printf("Error writing temp file: %v\n", err)
		return
	}
	_ = tmpfile.Close()

	// Set some environment variables (for example purposes only)
	_ = os.Setenv("REDIS_PORT", "6380")
	_ = os.Setenv("API_TIMEOUT", "45s")

	// STEP2: Create and load configuration
	cfg := &AppConfig{}
	if err := conf.Load(cfg, tmpfile.Name()); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Print the resulting configuration
	fmt.Printf("Environment: %s\n", cfg.App.Environment)
	fmt.Printf("Redis Host: %s\n", cfg.Redis.Host)
	fmt.Printf("Redis Port: %d\n", cfg.Redis.Port)
	fmt.Printf("Redis Timeout: %s\n", cfg.Redis.Timeout)
	fmt.Printf("API Endpoint: %s\n", cfg.API.Endpoint)
	fmt.Printf("API Timeout: %s\n", cfg.API.Timeout)

	// Output:
	// Environment: production
	// Redis Host: redis.prod.example.com
	// Redis Port: 6380
	// Redis Timeout: 10s
	// API Endpoint: https://api.prod.example.com
	// API Timeout: 45s
}

// ExampleDuration demonstrates how to use the Duration type
func ExampleDuration() {
	type Config struct {
		Timeout conf.Duration `json:"timeout" env:"TIMEOUT" default:"30s"`
	}

	cfg := &Config{}

	// Load with just defaults
	if err := conf.Load(cfg); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Default timeout: %s\n", cfg.Timeout)

	// Set environment variable
	_ = os.Setenv("TIMEOUT", "1m30s")

	// Load again with environment variable
	if err := conf.Load(cfg); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Environment timeout: %s\n", cfg.Timeout)

	// Output:
	// Default timeout: 30s
	// Environment timeout: 1m30s
}

func ExampleBasicFlags() {
	// STEP1: Define a custom configuration struct
	type AppConfig struct {
		conf.BaseConfig
		API struct {
			Endpoint string        `json:"endpoint" env:"API_ENDPOINT" default:"http://api.local"`
			Timeout  conf.Duration `json:"timeout" env:"API_TIMEOUT" default:"30s"`
		} `json:"api"`
	}

	// Create a temporary config file (for example purposes only)
	configJSON := `{
        "api": {
            "endpoint": "https://api.example.com"
        }
    }`
	tmpfile, err := os.CreateTemp("", "config.*.json")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configJSON)); err != nil {
		fmt.Printf("Error writing temp file: %v\n", err)
		return
	}
	_ = tmpfile.Close()

	// Save and restore os.Args (for example purposes only)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Simulate command line arguments
	os.Args = []string{
		"myapp",
		"--config", tmpfile.Name(),
		"--env", "production",
		"--debug",
		"--api-timeout", "1m",
	}

	// STEP2: Set up flags
	fs := flag.NewFlagSet("myapp", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Disable flag output for example

	// STEP3: Add basic flags
	conf.BasicFlags(fs)

	// STEP4: Add custom flags for this application
	apiTimeout := fs.Duration("api-timeout", 0, "API timeout duration")

	// STEP5: Parse flags
	_ = fs.Parse(os.Args[1:])

	// STEP6: Create and load configuration
	cfg := &AppConfig{}
	if err := conf.Load(cfg, fs.Lookup("config").Value.String()); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// STEP7: Apply flag overrides
	conf.ApplyFlagOverrides(&cfg.BaseConfig, fs)

	// Apply other flag values that should override config
	if apiTimeout != nil && *apiTimeout != 0 {
		cfg.API.Timeout = conf.Duration{Duration: *apiTimeout}
	}

	// Print the resulting configuration
	fmt.Printf("Environment: %s\n", cfg.App.Environment)
	fmt.Printf("Debug: %v\n", cfg.App.Debug)
	fmt.Printf("API Endpoint: %s\n", cfg.API.Endpoint)
	fmt.Printf("API Timeout: %s\n", cfg.API.Timeout)

	// Output:
	// Environment: production
	// Debug: true
	// API Endpoint: https://api.example.com
	// API Timeout: 1m0s
}
