package env_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/env"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		name         string
		setup        func()
		key          string
		defaultValue env.LogLevel
		expect       env.LogLevel
	}{
		{
			name: "log level debug",
			setup: func() {
				_ = os.Setenv("TEST_LOGLEVEL", "debug")
			},
			key:          "LOGLEVEL",
			defaultValue: env.LogInfo,
			expect:       env.LogDebug,
		},
		{
			name: "log level info",
			setup: func() {
				_ = os.Setenv("TEST_LOGLEVEL", "info")
			},
			key:          "LOGLEVEL",
			defaultValue: env.LogWarn,
			expect:       env.LogInfo,
		},
		{
			name: "log level warn",
			setup: func() {
				_ = os.Setenv("TEST_LOGLEVEL", "warn")
			},
			key:          "LOGLEVEL",
			defaultValue: env.LogError,
			expect:       env.LogWarn,
		},
		{
			name: "log level error",
			setup: func() {
				_ = os.Setenv("TEST_LOGLEVEL", "error")
			},
			key:          "LOGLEVEL",
			defaultValue: env.LogDebug,
			expect:       env.LogError,
		},
		{
			name:         "log level default",
			key:          "MISSING_LOGLEVEL",
			defaultValue: env.LogInfo,
			expect:       env.LogInfo,
		},
		{
			name: "invalid log level returns default",
			setup: func() {
				_ = os.Setenv("TEST_INVALID_LOGLEVEL", "invalid")
			},
			key:          "INVALID_LOGLEVEL",
			defaultValue: env.LogWarn,
			expect:       env.LogWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			t.Cleanup(func() {
				if tt.key != "" {
					_ = os.Unsetenv("TEST_" + tt.key)
				}
			})

			cfg := env.New("TEST_")
			got := cfg.LogLevel(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expect, got)
		})
	}
}

func TestMustLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		key         string
		expect      env.LogLevel
		expectPanic bool
	}{
		{
			name: "must log level debug",
			setup: func() {
				_ = os.Setenv("TEST_MUST_LOGLEVEL", "debug")
			},
			key:    "MUST_LOGLEVEL",
			expect: env.LogDebug,
		},
		{
			name: "must log level info",
			setup: func() {
				_ = os.Setenv("TEST_MUST_LOGLEVEL", "info")
			},
			key:    "MUST_LOGLEVEL",
			expect: env.LogInfo,
		},
		{
			name: "must log level warn",
			setup: func() {
				_ = os.Setenv("TEST_MUST_LOGLEVEL", "warn")
			},
			key:    "MUST_LOGLEVEL",
			expect: env.LogWarn,
		},
		{
			name: "must log level error",
			setup: func() {
				_ = os.Setenv("TEST_MUST_LOGLEVEL", "error")
			},
			key:    "MUST_LOGLEVEL",
			expect: env.LogError,
		},
		{
			name:        "must log level missing panics",
			key:         "MISSING_MUST_LOGLEVEL",
			expectPanic: true,
		},
		{
			name: "invalid must log level panics",
			setup: func() {
				_ = os.Setenv("TEST_INVALID_MUST_LOGLEVEL", "invalid")
			},
			key:         "INVALID_MUST_LOGLEVEL",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			t.Cleanup(func() {
				if tt.key != "" {
					_ = os.Unsetenv("TEST_" + tt.key)
				}
			})

			cfg := env.New("TEST_")

			if tt.expectPanic {
				assert.Panics(t, func() {
					cfg.MustLogLevel(tt.key)
				})
				return
			}

			got := cfg.MustLogLevel(tt.key)
			assert.Equal(t, tt.expect, got)
		})
	}
}
