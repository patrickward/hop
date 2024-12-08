// Package conf provides a way to load configuration from JSON files and environment variables,
// along with a structure to hold the configuration settings for an application and the ability
// to set up command-line flags for configuration options.
package conf

// HopConfig provides core configuration options
type HopConfig struct {
	App         AppConfig         `json:"app"`
	Csrf        CsrfConfig        `json:"csrf"`
	Events      EventsConfig      `json:"events"`
	Log         LogConfig         `json:"log"`
	Maintenance MaintenanceConfig `json:"maintenance"`
	Server      ServerConfig      `json:"server"`
	Session     SessionConfig     `json:"session"`
}

func (c *HopConfig) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *HopConfig) IsProduction() bool {
	return c.App.Environment == "production"
}

type AppConfig struct {
	Environment string `json:"environment" default:"development"`
	Debug       bool   `json:"debug" default:"false"`
}

type EventsConfig struct {
	MaxHistory int  `json:"max_history" default:"100"`
	DebugMode  bool `json:"debug_mode" default:"false"`
}

type LogConfig struct {
	Format      string `json:"format" default:"pretty"`
	IncludeTime bool   `json:"include_time" default:"false"`
	Level       string `json:"level" default:"debug"`
	Verbose     bool   `json:"verbose" default:"false"`
}

type MaintenanceConfig struct {
	Enabled bool   `json:"enabled" default:"false"`
	Message string `json:"message" default:""`
}

type CsrfConfig struct {
	HTTPOnly bool   `json:"http_only" default:"true"`
	Path     string `json:"path" default:"/"`
	MaxAge   int    `json:"max_age" default:"86400"`
	SameSite string `json:"same_site" default:"Lax"`
	Secure   bool   `json:"secure" default:"true"`
}

type SessionConfig struct {
	Lifetime      Duration `json:"lifetime" default:"168h"`
	CookiePersist bool     `json:"cookie_persist" default:"true"`
	// Other same-site values: "none", "strict"
	CookieSameSite string `json:"cookie_same_site" default:"lax"`
	CookieSecure   bool   `json:"cookie_secure" default:"true"`
	CookieHTTPOnly bool   `json:"cookie_http_only" default:"true"`
	CookiePath     string `json:"cookie_path" default:"/"`
}

type ServerConfig struct {
	BaseURL         string   `json:"base_url" default:"http://localhost:4444"`
	Host            string   `json:"host" default:"localhost"`
	Port            int      `json:"port" default:"4444"`
	IdleTimeout     Duration `json:"idle_timeout" default:"120s"`
	ReadTimeout     Duration `json:"read_timeout" default:"15s"`
	WriteTimeout    Duration `json:"write_timeout" default:"15s"`
	ShutdownTimeout Duration `json:"shutdown_timeout" default:"10s"`
}
