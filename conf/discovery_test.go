package conf_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/conf"
)

type DiscoveryConfig struct {
	Hop    conf.HopConfig
	Server struct {
		Host string `json:"host" default:"localhost"`
		Port int    `json:"port" default:"8080"`
	} `json:"server"`
}

func TestConfigDiscovery(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	// Create config directory
	configDir := filepath.Join(tmpDir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	// Helper to create config files with content
	createConfigFile := func(path, content string) {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	tests := []struct {
		name        string
		files       map[string]string // map of relative path to content
		environment string
		want        DiscoveryConfig
	}{
		{
			name: "base_config_only",
			files: map[string]string{
				"config.json": `{
					"server": {
						"host": "example.com",
						"port": 9000
					}
				}`,
			},
			want: DiscoveryConfig{
				Server: struct {
					Host string `json:"host" default:"localhost"`
					Port int    `json:"port" default:"8080"`
				}{
					Host: "example.com",
					Port: 9000,
				},
			},
		},
		{
			name: "environment_override",
			files: map[string]string{
				"config.json": `{
					"server": {
						"host": "example.com",
						"port": 9000
					}
				}`,
				"config/config.development.json": `{
					"server": {
						"port": 3000
					}
				}`,
			},
			environment: "development",
			want: DiscoveryConfig{
				Server: struct {
					Host string `json:"host" default:"localhost"`
					Port int    `json:"port" default:"8080"`
				}{
					Host: "example.com",
					Port: 3000,
				},
			},
		},
		{
			name: "local_override",
			files: map[string]string{
				"config.json": `{
					"server": {
						"host": "example.com",
						"port": 9000
					}
				}`,
				"config.local.json": `{
					"server": {
						"host": "localhost",
						"port": 8080
					}
				}`,
			},
			want: DiscoveryConfig{
				Server: struct {
					Host string `json:"host" default:"localhost"`
					Port int    `json:"port" default:"8080"`
				}{
					Host: "localhost",
					Port: 8080,
				},
			},
		},
		{
			name: "full_override_chain",
			files: map[string]string{
				"config.json": `{
					"server": {
						"host": "example.com",
						"port": 9000
					}
				}`,
				"config/config.json": `{
					"server": {
						"port": 9001
					}
				}`,
				"config.development.json": `{
					"server": {
						"port": 3000
					}
				}`,
				"config.local.json": `{
					"server": {
						"host": "localhost"
					}
				}`,
			},
			environment: "development",
			want: DiscoveryConfig{
				Server: struct {
					Host string `json:"host" default:"localhost"`
					Port int    `json:"port" default:"8080"`
				}{
					Host: "localhost",
					Port: 3000,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test files
			for path, content := range tt.files {
				createConfigFile(path, content)
			}

			// Change to temp directory for test
			originalWd, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))
			defer func(dir string) {
				_ = os.Chdir(dir)
			}(originalWd)

			// Create config and manager
			cfg := &DiscoveryConfig{}
			var mgr *conf.Manager
			if tt.environment != "" {
				mgr = conf.NewManager(cfg, conf.WithEnvironment(tt.environment))
			} else {
				mgr = conf.NewManager(cfg)
			}

			// Load configuration
			err = mgr.Load()
			require.NoError(t, err)

			// Verify configuration
			assert.Equal(t, tt.want.Server.Host, cfg.Server.Host)
			assert.Equal(t, tt.want.Server.Port, cfg.Server.Port)

			// Clean up test files
			for path := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				_ = os.Remove(fullPath)
			}
		})
	}
}

// TestMissingConfigs ensures the system works when no config files exist
func TestMissingConfigs(t *testing.T) {
	cfg := &DiscoveryConfig{}
	mgr := conf.NewManager(cfg)

	err := mgr.Load()
	require.NoError(t, err)

	// Should have default values
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
}

// TestInvalidJSON ensures proper error handling for malformed config files
func TestInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tmpDir)

	// Create invalid JSON file
	configPath := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{invalid json}`), 0644))

	require.NoError(t, os.Chdir(tmpDir))

	cfg := &DiscoveryConfig{}
	mgr := conf.NewManager(cfg)

	err = mgr.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
