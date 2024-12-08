package conf_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf"
)

type TestConfig struct {
	Hop conf.HopConfig
	API struct {
		Endpoint    string        `json:"endpoint" default:"http://api.local"`
		Timeout     conf.Duration `json:"timeout" default:"30s"`
		MaxRetries  int           `json:"max_retries" default:"3"`
		RetryDelay  conf.Duration `json:"retry_delay" default:"5s"`
		EnableCache bool          `json:"enable_cache" default:"true"`
	} `json:"api"`
}

func TestConfigManager(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		env         map[string]string
		validate    func(*testing.T, *TestConfig)
		expectError bool
	}{
		{
			name:  "defaults_only",
			files: []string{},
			env:   map[string]string{},
			validate: func(t *testing.T, cfg *TestConfig) {
				assertion := assert.New(t)
				assertion.Equal("http://api.local", cfg.API.Endpoint, "default endpoint should be set")
				assertion.Equal(30*time.Second, cfg.API.Timeout.Duration, "default timeout should be set")
				assertion.Equal(3, cfg.API.MaxRetries, "default max retries should be set")
				assertion.True(cfg.API.EnableCache, "default enable cache should be true")
			},
		},
		{
			name:  "load_from_json",
			files: []string{"testdata/config.json"},
			env:   map[string]string{},
			validate: func(t *testing.T, cfg *TestConfig) {
				assertion := assert.New(t)
				assertion.Equal("example.com", cfg.Hop.Server.Host, "host should be loaded from json")
				assertion.Equal(9000, cfg.Hop.Server.Port, "port should be loaded from json")
				assertion.Equal(time.Minute, cfg.API.Timeout.Duration, "timeout should be loaded from json")
			},
		},
		{
			name:  "env_overrides_json",
			files: []string{"testdata/config.json"},
			env: map[string]string{
				"HOP_SERVER_PORT": "8080",
				"API_ENDPOINT":    "https://override.com",
				"API_TIMEOUT":     "45s",
				"API_MAX_RETRIES": "5",
			},
			validate: func(t *testing.T, cfg *TestConfig) {
				assertion := assert.New(t)
				assertion.Equal(8080, cfg.Hop.Server.Port, "port should be overridden by env")
				assertion.Equal("https://override.com", cfg.API.Endpoint, "endpoint should be overridden by env")
				assertion.Equal(45*time.Second, cfg.API.Timeout.Duration, "timeout should be overridden by env")
				assertion.Equal(5, cfg.API.MaxRetries, "max retries should be overridden by env")
			},
		},
		{
			name:  "duration_parsing",
			files: []string{},
			env: map[string]string{
				"API_TIMEOUT":     "2h30m",
				"API_RETRY_DELAY": "250ms",
			},
			validate: func(t *testing.T, cfg *TestConfig) {
				assertion := assert.New(t)
				expectedTimeout := 2*time.Hour + 30*time.Minute
				assertion.Equal(expectedTimeout, cfg.API.Timeout.Duration, "timeout should be parsed correctly")
				assertion.Equal(250*time.Millisecond, cfg.API.RetryDelay.Duration, "retry delay should be parsed correctly")
			},
		},
		{
			name:  "invalid_duration",
			files: []string{},
			env: map[string]string{
				"API_TIMEOUT": "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment first
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.env {
				require.NoError(t, os.Setenv(k, v), "setting env variable %s=%s", k, v)
			}

			cfg := &TestConfig{}
			mgr := conf.NewManager(cfg, conf.WithConfigFiles(tt.files...))
			err := mgr.Load()

			if tt.expectError {
				assert.Error(t, err, "should return error")
				return
			}

			require.NoError(t, err, "should load without error")
			tt.validate(t, cfg)
		})
	}
}

func TestConfigManager_WithOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []conf.Option
		env      map[string]string
		validate func(*testing.T, *TestConfig)
	}{
		{
			name: "with_env_prefix",
			options: []conf.Option{
				conf.WithEnvPrefix("APP"),
			},
			env: map[string]string{
				"APP_API_ENDPOINT": "https://prefixed.com",
			},
			validate: func(t *testing.T, cfg *TestConfig) {
				assert.Equal(t, "https://prefixed.com", cfg.API.Endpoint)
			},
		},
		// Add more option tests as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.env {
				require.NoError(t, os.Setenv(k, v))
			}

			cfg := &TestConfig{}
			mgr := conf.NewManager(cfg, tt.options...)
			require.NoError(t, mgr.Load())
			tt.validate(t, cfg)
		})
	}
}

func TestConfigManager_Reload(t *testing.T) {
	os.Clearenv()

	cfg := &TestConfig{}
	mgr := conf.NewManager(cfg)

	// Initial load
	require.NoError(t, mgr.Load())
	assert.Equal(t, "http://api.local", cfg.API.Endpoint) // default value

	// Change env and reload
	_ = os.Setenv("API_ENDPOINT", "https://reloaded.com")
	require.NoError(t, mgr.Reload())

	cfg = mgr.Get().(*TestConfig)
	assert.Equal(t, "https://reloaded.com", cfg.API.Endpoint)

	os.Clearenv()
}

// Let's also add a test specifically for handling invalid configuration during reload
func TestConfigManager_ReloadWithInvalidConfig(t *testing.T) {
	os.Clearenv()

	cfg := &TestConfig{}
	mgr := conf.NewManager(cfg)

	// Initial load should succeed
	require.NoError(t, mgr.Load(), "initial load should succeed")
	initialEndpoint := cfg.API.Endpoint
	initialTimeout := cfg.API.Timeout

	// Set invalid configuration
	err := os.Setenv("API_TIMEOUT", "invalid")
	require.NoError(t, err, "setting invalid env var")

	// Reload should fail
	err = mgr.Reload()
	assert.Error(t, err, "reload should fail with invalid configuration")
	assert.Contains(t, err.Error(), "invalid duration")

	// Original configuration should be unchanged
	assert.Equal(t, initialEndpoint, cfg.API.Endpoint, "config should retain original endpoint")
	assert.Equal(t, initialTimeout, cfg.API.Timeout, "config should retain original timeout")

	// Clean up
	os.Clearenv()
}
