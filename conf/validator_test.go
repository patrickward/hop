package conf_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf"
	"github.com/patrickward/hop/conf/conftype"
)

// Test types for validation
type ValidConfig struct {
	Hop conf.HopConfig `json:"hop"`
	API struct {
		Timeout    conftype.Duration `json:"timeout" default:"30s"`
		MaxRetries int               `json:"max_retries" default:"3"`
	} `json:"api"`
}

func (c *ValidConfig) Validate() error {
	if c.API.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	return nil
}

type InvalidConfig struct {
	// Missing Hop framework config
	API struct {
		Timeout conftype.Duration `json:"timeout"`
	} `json:"api"`
}

type InvalidCustomConfig struct {
	Hop conf.HopConfig `json:"hop"`
	API struct {
		MaxRetries int `json:"max_retries" default:"-1"` // Invalid default
	} `json:"api"`
}

func (c *InvalidCustomConfig) Validate() error {
	if c.API.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	return nil
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      interface{}
		env         map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_config",
			config:      &ValidConfig{},
			env:         map[string]string{},
			expectError: false,
		},
		{
			name:        "missing_hop_config",
			config:      &InvalidConfig{},
			env:         map[string]string{},
			expectError: true,
			errorMsg:    "must include Hop framework configuration",
		},
		{
			name:        "invalid_custom_validation",
			config:      &InvalidCustomConfig{},
			env:         map[string]string{},
			expectError: true,
			errorMsg:    "max retries cannot be negative",
		},
		{
			name:   "invalid_env_override",
			config: &ValidConfig{},
			env: map[string]string{
				"API_MAX_RETRIES": "-5",
			},
			expectError: true,
			errorMsg:    "max retries cannot be negative",
		},
		{
			name:   "valid_env_override",
			config: &ValidConfig{},
			env: map[string]string{
				"API_MAX_RETRIES": "5",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.env {
				require.NoError(t, os.Setenv(k, v))
			}

			// Create manager and load config
			mgr := conf.NewManager(tt.config)
			err := mgr.Load()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			assert.NoError(t, err)

			// If validation passes, verify we can reload
			if err := mgr.Reload(); err != nil {
				t.Errorf("unexpected error on reload: %v", err)
			}
		})
	}
}
