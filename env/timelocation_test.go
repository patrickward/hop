package env_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/env"
)

func TestTimeLocation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		key         string
		method      func(cfg *env.Config) any
		expect      any
		expectPanic bool
	}{
		{
			name: "time location present",
			setup: func() {
				_ = os.Setenv("TEST_TIMEZONE", "America/New_York")
			},
			key: "TIMEZONE",
			method: func(cfg *env.Config) any {
				return cfg.TimeLocation("TIMEZONE", time.UTC)
			},
			expect: func() *time.Location {
				loc, _ := time.LoadLocation("America/New_York")
				return loc
			}(),
		},
		{
			name: "time location invalid returns default",
			setup: func() {
				_ = os.Setenv("TEST_TIMEZONE_INVALID", "Invalid/Timezone")
			},
			key: "TIMEZONE_INVALID",
			method: func(cfg *env.Config) any {
				return cfg.TimeLocation("TIMEZONE_INVALID", time.UTC)
			},
			expect: time.UTC,
		},
		{
			name: "time location not set returns default",
			key:  "TIMEZONE_NOT_SET",
			method: func(cfg *env.Config) any {
				return cfg.TimeLocation("TIMEZONE_NOT_SET", time.UTC)
			},
			expect: time.UTC,
		},
		{
			name: "must time location present",
			setup: func() {
				_ = os.Setenv("TEST_MUST_TIMEZONE", "Europe/London")
			},
			key: "MUST_TIMEZONE",
			method: func(cfg *env.Config) any {
				return cfg.MustTimeLocation("MUST_TIMEZONE")
			},
			expect: func() *time.Location {
				loc, _ := time.LoadLocation("Europe/London")
				return loc
			}(),
		},
		{
			name: "must time location missing panics",
			key:  "MUST_TIMEZONE_MISSING",
			method: func(cfg *env.Config) any {
				return cfg.MustTimeLocation("MUST_TIMEZONE_MISSING")
			},
			expectPanic: true,
		},
		{
			name: "must time location invalid panics",
			setup: func() {
				_ = os.Setenv("TEST_MUST_TIMEZONE_INVALID", "Invalid/Timezone")
			},
			key: "MUST_TIMEZONE_INVALID",
			method: func(cfg *env.Config) any {
				return cfg.MustTimeLocation("MUST_TIMEZONE_INVALID")
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
