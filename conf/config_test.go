package conf_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/patrickward/hop/conf"
)

type TestConfig struct {
	conf.Config
	API struct {
		Endpoint    string        `json:"endpoint" env:"API_ENDPOINT" default:"http://api.local"`
		Timeout     conf.Duration `json:"timeout" env:"API_TIMEOUT" default:"30s"`
		MaxRetries  int           `json:"max_retries" env:"API_MAX_RETRIES" default:"3"`
		RetryDelay  conf.Duration `json:"retry_delay" env:"API_RETRY_DELAY" default:"5s"`
		EnableCache bool          `json:"enable_cache" env:"API_ENABLE_CACHE" default:"true"`
	} `json:"api"`
}

func TestConfigLoading(t *testing.T) {
	// Test cases for different loading scenarios
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
				if cfg.API.Endpoint != "http://api.local" {
					t.Errorf("expected default endpoint http://api.local, got %s", cfg.API.Endpoint)
				}
				if cfg.API.Timeout.Duration != 30*time.Second {
					t.Errorf("expected default timeout 30s, got %s", cfg.API.Timeout)
				}
				if cfg.API.MaxRetries != 3 {
					t.Errorf("expected default max retries 3, got %d", cfg.API.MaxRetries)
				}
				if !cfg.API.EnableCache {
					t.Error("expected default enable cache to be true")
				}
			},
		},
		{
			name:  "load_from_json",
			files: []string{"testdata/config.json"},
			env:   map[string]string{},
			validate: func(t *testing.T, cfg *TestConfig) {
				if cfg.Server.Host != "example.com" {
					t.Errorf("expected host example.com, got %s", cfg.Server.Host)
				}
				if cfg.Server.Port != 9000 {
					t.Errorf("expected port 9000, got %d", cfg.Server.Port)
				}
				if cfg.API.Timeout.Duration != time.Minute {
					t.Errorf("expected timeout 1m, got %s", cfg.API.Timeout)
				}
			},
		},
		{
			name:  "env_overrides_json",
			files: []string{"testdata/config.json"},
			env: map[string]string{
				"SERVER_PORT":     "8080",
				"API_ENDPOINT":    "https://override.com",
				"API_TIMEOUT":     "45s",
				"API_MAX_RETRIES": "5",
			},
			validate: func(t *testing.T, cfg *TestConfig) {
				if cfg.Server.Port != 8080 {
					t.Errorf("expected port 8080, got %d", cfg.Server.Port)
				}
				if cfg.API.Endpoint != "https://override.com" {
					t.Errorf("expected endpoint https://override.com, got %s", cfg.API.Endpoint)
				}
				if cfg.API.Timeout.Duration != 45*time.Second {
					t.Errorf("expected timeout 45s, got %s", cfg.API.Timeout)
				}
				if cfg.API.MaxRetries != 5 {
					t.Errorf("expected max retries 5, got %d", cfg.API.MaxRetries)
				}
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
				expected := 2*time.Hour + 30*time.Minute
				if cfg.API.Timeout.Duration != expected {
					t.Errorf("expected timeout %s, got %s", expected, cfg.API.Timeout)
				}
				if cfg.API.RetryDelay.Duration != 250*time.Millisecond {
					t.Errorf("expected retry delay 250ms, got %s", cfg.API.RetryDelay)
				}
			},
		},
		{
			name:  "invalid_duration",
			files: []string{},
			env: map[string]string{
				"API_TIMEOUT": "invalid",
			},
			expectError: true,                                   // Now expecting this to error
			validate:    func(t *testing.T, cfg *TestConfig) {}, // No validation needed as we expect error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment first
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.env {
				err := os.Setenv(k, v)
				if err != nil {
					t.Fatalf("failed to set env variable %s: %v", k, err)
				}
			}

			cfg := &TestConfig{}

			// Load config
			err := conf.Load(cfg, tt.files...)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return // Skip validation if we expect an error
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Run validation
			tt.validate(t, cfg)
		})
	}
}

func TestDurationJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{`"1h"`, time.Hour, false},
		{`"30s"`, 30 * time.Second, false},
		{`"2h30m"`, 2*time.Hour + 30*time.Minute, false},
		{`"invalid"`, 0, true},
		{`""`, 0, true},
	}

	for _, tt := range tests {
		var d conf.Duration
		err := json.Unmarshal([]byte(tt.input), &d)

		if tt.wantErr {
			if err == nil {
				t.Errorf("expected error for input %s", tt.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("unexpected error for input %s: %v", tt.input, err)
			continue
		}

		if d.Duration != tt.expected {
			t.Errorf("for input %s: expected %s, got %s", tt.input, tt.expected, d.Duration)
		}
	}
}
