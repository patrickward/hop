// Package conf provides a way to load configuration from JSON files and environment variables,
// along with a structure to hold the configuration settings for an application and the ability
// to set up command-line flags for configuration options.
package conf

import "github.com/patrickward/hop/conf/conftype"

// HopConfig provides core configuration options
type HopConfig struct {
	App         AppConfig         `json:"app" koanf:"app"`
	Csrf        CsrfConfig        `json:"csrf" koanf:"csrf"`
	Events      EventsConfig      `json:"events" koanf:"events"`
	Log         LogConfig         `json:"log" koanf:"log"`
	Maintenance MaintenanceConfig `json:"maintenance" koanf:"maintenance"`
	Server      ServerConfig      `json:"server" koanf:"server"`
	Session     SessionConfig     `json:"session" koanf:"session"`
}

func (c *HopConfig) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *HopConfig) IsProduction() bool {
	return c.App.Environment == "production"
}

type AppConfig struct {
	Environment string `json:"environment" default:"development" koanf:"environment"`
	Debug       bool   `json:"debug" default:"false" koanf:"debug"`
}

type EventsConfig struct {
	MaxHistory int  `json:"max_history" default:"100" koanf:"max_history"`
	DebugMode  bool `json:"debug_mode" default:"false" koanf:"debug_mode"`
}

type LogConfig struct {
	Format      string `json:"format" default:"pretty" koanf:"format"`
	IncludeTime bool   `json:"include_time" default:"false" koanf:"include_time"`
	Level       string `json:"level" default:"debug" koanf:"level"`
	Verbose     bool   `json:"verbose" default:"false" koanf:"verbose"`
}

type MaintenanceConfig struct {
	Enabled bool   `json:"enabled" default:"false" koanf:"enabled"`
	Message string `json:"message" default:"" koanf:"message"`
}

type CsrfConfig struct {
	HTTPOnly bool   `json:"http_only" default:"true" koanf:"http_only"`
	Path     string `json:"path" default:"/" koanf:"path"`
	MaxAge   int    `json:"max_age" default:"86400" koanf:"max_age"`
	SameSite string `json:"same_site" default:"Lax" koanf:"same_site"`
	Secure   bool   `json:"secure" default:"true" koanf:"secure"`
}

type SessionConfig struct {
	Lifetime      conftype.Duration `json:"lifetime" default:"168h" koanf:"lifetime"`
	CookiePersist bool              `json:"cookie_persist" default:"true" koanf:"cookie_persist"`
	// Other same-site values: "none", "strict"
	CookieSameSite string `json:"cookie_same_site" default:"lax" koanf:"cookie_same_site"`
	CookieSecure   bool   `json:"cookie_secure" default:"true" koanf:"cookie_secure"`
	CookieHTTPOnly bool   `json:"cookie_http_only" default:"true" koanf:"cookie_http_only"`
	CookiePath     string `json:"cookie_path" default:"/" koanf:"cookie_path"`
}

type ServerConfig struct {
	BaseURL         string            `json:"base_url" default:"http://localhost:4444" koanf:"base_url"`
	Host            string            `json:"host" default:"localhost" koanf:"host"`
	Port            int               `json:"port" default:"4444" koanf:"port"`
	IdleTimeout     conftype.Duration `json:"idle_timeout" default:"120s" koanf:"idle_timeout"`
	ReadTimeout     conftype.Duration `json:"read_timeout" default:"15s" koanf:"read_timeout"`
	WriteTimeout    conftype.Duration `json:"write_timeout" default:"15s" koanf:"write_timeout"`
	ShutdownTimeout conftype.Duration `json:"shutdown_timeout" default:"10s" koanf:"shutdown_timeout"`
}
