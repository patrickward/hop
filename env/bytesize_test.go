package env_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/env"
)

func TestConfig_ByteSize(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		key         string
		method      func(cfg *env.Config) any
		expect      any
		expectPanic bool
	}{
		{
			name: "byte size present",
			setup: func() {
				_ = os.Setenv("TEST_BYTESIZE", "1GB")
			},
			key: "BYTESIZE",
			method: func(cfg *env.Config) any {
				return cfg.GetByteSize("BYTESIZE", 0)
			},
			expect: 1 << 30,
		},
		{
			name: "byte size MB",
			setup: func() {
				_ = os.Setenv("TEST_BYTESIZE_MB", "512MB")
			},
			key: "BYTESIZE_MB",
			method: func(cfg *env.Config) any {
				return cfg.GetByteSize("BYTESIZE_MB", 0)
			},
			expect: 512 << 20,
		},
		{
			name: "byte size KB",
			setup: func() {
				_ = os.Setenv("TEST_BYTESIZE_KB", "256KB")
			},
			key: "BYTESIZE_KB",
			method: func(cfg *env.Config) any {
				return cfg.GetByteSize("BYTESIZE_KB", 0)
			},
			expect: 256 << 10,
		},
		{
			name: "byte size B",
			setup: func() {
				_ = os.Setenv("TEST_BYTESIZE_B", "128B")
			},
			key: "BYTESIZE_B",
			method: func(cfg *env.Config) any {
				return cfg.GetByteSize("BYTESIZE_B", 0)
			},
			expect: 128,
		},
		{
			name: "byte size no suffix",
			setup: func() {
				_ = os.Setenv("TEST_BYTESIZE_NO_SUFFIX", "64")
			},
			key: "BYTESIZE_NO_SUFFIX",
			method: func(cfg *env.Config) any {
				return cfg.GetByteSize("BYTESIZE_NO_SUFFIX", 0)
			},
			expect: 64,
		},
		{
			name: "byte size lowercase (ok)",
			setup: func() {
				_ = os.Setenv("TEST_BYTESIZE_LOWERCASE", "2gb")
			},
			key: "BYTESIZE_LOWERCASE",
			method: func(cfg *env.Config) any {
				return cfg.GetByteSize("BYTESIZE_LOWERCASE", 0)
			},
			expect: 2 << 30,
		},
		{
			name: "byte size default",
			key:  "MISSING_BYTESIZE",
			method: func(cfg *env.Config) any {
				return cfg.GetByteSize("MISSING_BYTESIZE", 1024)
			},
			expect: 1024,
		},
		{
			name: "must byte size present",
			setup: func() {
				_ = os.Setenv("TEST_MUST_BYTESIZE", "512MB")
			},
			key: "MUST_BYTESIZE",
			method: func(cfg *env.Config) any {
				return cfg.MustByteSize("MUST_BYTESIZE")
			},
			expect: 512 << 20,
		},
		{
			name: "must byte size missing panics",
			key:  "MISSING_MUST_BYTESIZE",
			method: func(cfg *env.Config) any {
				return cfg.MustByteSize("MISSING_MUST_BYTESIZE")
			},
			expectPanic: true,
		},
		{
			name: "invalid byte size returns default",
			setup: func() {
				_ = os.Setenv("TEST_INVALID_BYTESIZE", "not-a-size")
			},
			key: "INVALID_BYTESIZE",
			method: func(cfg *env.Config) any {
				return cfg.GetByteSize("INVALID_BYTESIZE", 2048)
			},
			expect: 2048,
		},
		{
			name: "invalid must byte size panics",
			setup: func() {
				_ = os.Setenv("TEST_INVALID_MUST_BYTESIZE", "not-a-size")
			},
			key: "INVALID_MUST_BYTESIZE",
			method: func(cfg *env.Config) any {
				return cfg.MustByteSize("INVALID_MUST_BYTESIZE")
			},
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
					tt.method(cfg)
				})
				return
			}

			got := tt.method(cfg)
			assert.Equal(t, tt.expect, got)
		})
	}
}
