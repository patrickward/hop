package env_test

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/env"
)

func TestConfig_BasicTypes(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		key         string // Key without prefix
		method      func(cfg *env.Config) any
		expect      any
		expectPanic bool
	}{
		{
			name: "string present",
			setup: func() {
				_ = os.Setenv("TEST_STR", "hello")
			},
			key: "STR",
			method: func(cfg *env.Config) any {
				return cfg.String("STR", "default")
			},
			expect: "hello",
		},
		{
			name: "string default",
			key:  "MISSING_STR",
			method: func(cfg *env.Config) any {
				return cfg.String("MISSING_STR", "default")
			},
			expect: "default",
		},
		{
			name: "must string present",
			setup: func() {
				_ = os.Setenv("TEST_MUST_STR", "hello")
			},
			key: "MUST_STR",
			method: func(cfg *env.Config) any {
				return cfg.MustString("MUST_STR")
			},
			expect: "hello",
		},
		{
			name: "must string missing panics",
			key:  "MISSING_MUST_STR",
			method: func(cfg *env.Config) any {
				return cfg.MustString("MISSING_MUST_STR")
			},
			expectPanic: true,
		},
		{
			name: "int present",
			setup: func() {
				_ = os.Setenv("TEST_INT", "42")
			},
			key: "INT",
			method: func(cfg *env.Config) any {
				return cfg.Int("INT", 0)
			},
			expect: 42,
		},
		{
			name: "int default",
			key:  "MISSING_INT",
			method: func(cfg *env.Config) any {
				return cfg.Int("MISSING_INT", 42)
			},
			expect: 42,
		},
		{
			name: "invalid int returns default",
			setup: func() {
				_ = os.Setenv("TEST_INVALID_INT", "not-a-number")
			},
			key: "INVALID_INT",
			method: func(cfg *env.Config) any {
				return cfg.Int("INVALID_INT", 42)
			},
			expect: 42,
		},
		{
			name: "bool true",
			setup: func() {
				_ = os.Setenv("TEST_BOOL", "true")
			},
			key: "BOOL",
			method: func(cfg *env.Config) any {
				return cfg.Bool("BOOL", false)
			},
			expect: true,
		},
		{
			name: "bool yes",
			setup: func() {
				_ = os.Setenv("TEST_BOOL_YES", "yes")
			},
			key: "BOOL_YES",
			method: func(cfg *env.Config) any {
				return cfg.Bool("BOOL_YES", false)
			},
			expect: true,
		},
		{
			name: "bool 1",
			setup: func() {
				_ = os.Setenv("TEST_BOOL_1", "1")
			},
			key: "BOOL_1",
			method: func(cfg *env.Config) any {
				return cfg.Bool("BOOL_1", false)
			},
			expect: true,
		},
		{
			name: "bool default",
			key:  "MISSING_BOOL",
			method: func(cfg *env.Config) any {
				return cfg.Bool("MISSING_BOOL", true)
			},
			expect: true,
		},
		{
			name: "duration present",
			setup: func() {
				_ = os.Setenv("TEST_DURATION", "5m")
			},
			key: "DURATION",
			method: func(cfg *env.Config) any {
				return cfg.Duration("DURATION", time.Second)
			},
			expect: 5 * time.Minute,
		},
		{
			name: "duration default",
			key:  "MISSING_DURATION",
			method: func(cfg *env.Config) any {
				return cfg.Duration("MISSING_DURATION", time.Minute)
			},
			expect: time.Minute,
		},
		{
			name: "string slice present",
			setup: func() {
				_ = os.Setenv("TEST_SLICE", "a,b,c")
			},
			key: "SLICE",
			method: func(cfg *env.Config) any {
				return cfg.StringSlice("SLICE", []string{"default"})
			},
			expect: []string{"a", "b", "c"},
		},
		{
			name: "string slice with spaces",
			setup: func() {
				_ = os.Setenv("TEST_SLICE_SPACES", "a, b,  c")
			},
			key: "SLICE_SPACES",
			method: func(cfg *env.Config) any {
				return cfg.StringSlice("SLICE_SPACES", []string{"default"})
			},
			expect: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				tt.setup()
			}

			// Cleanup
			t.Cleanup(func() {
				_ = os.Unsetenv("TEST_" + tt.key)
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

func TestConfig_CustomGetters(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		register    func(*env.Config)
		key         string
		getterName  string
		method      func(cfg *env.Config) any
		expect      any
		expectPanic bool
	}{
		{
			name: "url getter",
			setup: func() {
				_ = os.Setenv("TEST_URL", "https://example.com")
			},
			register: func(cfg *env.Config) {
				cfg.RegisterGetter("url", func(value string) (any, error) {
					return url.Parse(value)
				})
			},
			key:        "URL",
			getterName: "url",
			method: func(cfg *env.Config) any {
				u := cfg.MustGet("URL", "url").(*url.URL)
				return u.String()
			},
			expect: "https://example.com",
		},
		{
			name: "url getter with default",
			register: func(cfg *env.Config) {
				cfg.RegisterGetter("url", func(value string) (any, error) {
					return url.Parse(value)
				})
			},
			key:        "MISSING_URL",
			getterName: "url",
			method: func(cfg *env.Config) any {
				defaultURL, _ := url.Parse("https://default.com")
				u := cfg.Get("MISSING_URL", "url", defaultURL).(*url.URL)
				return u.String()
			},
			expect: "https://default.com",
		},
		{
			name: "unregistered getter panics",
			method: func(cfg *env.Config) any {
				return cfg.Get("KEY", "missing", nil)
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
			if tt.register != nil {
				tt.register(cfg)
			}

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

func TestConfig_Prefix(t *testing.T) {
	// Test that prefix is properly applied
	cfg := env.New("PREFIX_")

	_ = os.Setenv("PREFIX_TEST", "value")
	_ = os.Setenv("NO_PREFIX_TEST", "wrong")

	t.Cleanup(func() {
		_ = os.Unsetenv("PREFIX_TEST")
		_ = os.Unsetenv("NO_PREFIX_TEST")
	})

	assert.Equal(t, "value", cfg.String("TEST", "default"))
}
