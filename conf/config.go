// Package conf provides a way to load configuration from JSON files and environment variables,
// along with a structure to hold the configuration settings for an application and the ability
// to set up command-line flags for configuration options.
package conf

import (
	"fmt"
)

// HopConfig provides core configuration options
type HopConfig struct {
	App         AppConfig         `json:"app"`
	Company     CompanyConfig     `json:"company"`
	Events      EventsConfig      `json:"events"`
	Maintenance MaintenanceConfig `json:"maintenance"`
	Session     SessionConfig     `json:"session"`
	Csrf        CsrfConfig        `json:"csrf"`
	Server      ServerConfig      `json:"server"`
	Database    DatabaseConfig    `json:"database"`
	SMTP        SMPTConfig        `json:"smtp"`
	Log         LogConfig         `json:"log"`
}

func (c *HopConfig) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *HopConfig) IsProduction() bool {
	return c.App.Environment == "production"
}

type AppConfig struct {
	Environment string `json:"environment" env:"APP_ENVIRONMENT" default:"development"`
	Debug       bool   `json:"debug" env:"APP_DEBUG" default:"false"`
	SAASMode    bool   `json:"saas_mode" env:"SAAS_MODE" default:"false"`
}

type EventsConfig struct {
	MaxHistory int  `json:"max_history" env:"EVENTS_MAX_HISTORY" default:"100"`
	DebugMode  bool `json:"debug_mode" env:"EVENTS_DEBUG_MODE" default:"false"`
}

type MaintenanceConfig struct {
	Enabled bool   `json:"enabled" env:"MAINTENANCE_ENABLED" default:"false"`
	Message string `json:"message" env:"MAINTENANCE_MESSAGE" default:""`
}

type CsrfConfig struct {
	HTTPOnly bool   `json:"http_only" env:"CSRF_HTTP_ONLY" default:"true"`
	Path     string `json:"path" env:"CSRF_PATH" default:"/"`
	MaxAge   int    `json:"max_age" env:"CSRF_MAX_AGE" default:"86400"`
	SameSite string `json:"same_site" env:"CSRF_SAME_SITE" default:"Lax"`
	Secure   bool   `json:"secure" env:"CSRF_SECURE" default:"true"`
}

type SessionConfig struct {
	Lifetime      Duration `json:"lifetime" env:"SESSION_LIFETIME" default:"168h"`
	CookiePersist bool     `json:"cookie_persist" env:"SESSION_COOKIE_PERSIST" default:"true"`
	// Other same-site values: "None", "Strict"
	CookieSameSite string `json:"cookie_same_site" env:"SESSION_COOKIE_SAME_SITE" default:"Lax"`
	CookieSecure   bool   `json:"cookie_secure" env:"SESSION_COOKIE_SECURE" default:"true"`
	CookieHTTPOnly bool   `json:"cookie_http_only" env:"SESSION_COOKIE_HTTP_ONLY" default:"true"`
	CookiePath     string `json:"cookie_path" env:"SESSION_COOKIE_PATH" default:"/"`
}

type ServerConfig struct {
	BaseURL         string   `json:"base_url" env:"SERVER_BASE_URL" default:"http://localhost:4444"`
	Host            string   `json:"host" env:"SERVER_HOST" default:"localhost"`
	Port            int      `json:"port" env:"SERVER_PORT" default:"4444"`
	IdleTimeout     Duration `json:"idle_timeout" env:"SERVER_IDLE_TIMEOUT" default:"120s"`
	ReadTimeout     Duration `json:"read_timeout" env:"SERVER_READ_TIMEOUT" default:"15s"`
	WriteTimeout    Duration `json:"write_timeout" env:"SERVER_WRITE_TIMEOUT" default:"15s"`
	ShutdownTimeout Duration `json:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT" default:"10s"`
}

type DatabaseConfig struct {
	Driver          string   `json:"driver" env:"DB_DRIVER" default:"sqlite"`
	URI             string   `json:"uri" env:"DB_URI" default:"data/db.sqlite"`
	Timeout         Duration `json:"timeout" env:"DB_TIMEOUT" default:"10s"`
	MaxIdleConns    int      `json:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"10"`
	MaxIdleTime     Duration `json:"max_idle_time" env:"DB_MAX_IDLE_TIME" default:"5m"`
	MaxConnLifetime Duration `json:"max_conn_lifetime" env:"DB_MAX_CONN_LIFETIME" default:"30m"`
	AutoMigrate     bool     `json:"auto_migrate" env:"DB_AUTO_MIGRATE" default:"false"`
}

type SMPTConfig struct {
	Enabled    bool     `json:"enabled" env:"SMTP_ENABLED" default:"false"`
	Host       string   `json:"host" env:"SMTP_HOST" default:"localhost"`
	Port       int      `json:"port" env:"SMTP_PORT" default:"1025"`
	Username   string   `json:"username" env:"SMTP_USERNAME" default:""`
	Password   string   `json:"password" env:"SMTP_PASSWORD" default:""`
	From       string   `json:"from" env:"SMTP_FROM" default:""`
	AuthType   string   `json:"auth_type" env:"SMTP_AUTH_TYPE" default:"LOGIN"`
	TLSPolicy  int      `json:"tls_policy" env:"SMTP_TLS_POLICY" default:"1"`
	RetryCount int      `json:"retry_count" env:"SMTP_RETRY_COUNT" default:"3"`
	RetryDelay Duration `json:"retry_delay" env:"SMTP_RETRY_DELAY" default:"5s"`
}

type CompanyConfig struct {
	Address      string `json:"address" env:"COMPANY_ADDRESS" default:""`
	Address2     string `json:"address2" env:"COMPANY_ADDRESS2" default:""`
	City         string `json:"city" env:"COMPANY_CITY" default:""`
	State        string `json:"state" env:"COMPANY_STATE" default:""`
	Zip          string `json:"zip" env:"COMPANY_ZIP" default:""`
	Name         string `json:"name" env:"COMPANY_NAME" default:""`
	LogoURL      string `json:"logo_url" env:"COMPANY_LOGO_URL" default:""`
	SupportEmail string `json:"support_email" env:"COMPANY_SUPPORT_EMAIL" default:""`
	SupportPhone string `json:"support_phone" env:"COMPANY_SUPPORT_PHONE" default:""`
	WebsiteName  string `json:"website_name" env:"COMPANY_WEBSITE_NAME" default:""`
	WebsiteURL   string `json:"website_url" env:"COMPANY_WEBSITE_URL" default:""`
	//SiteLinks        map[string]string `json:"site_links" env:"SITE_LINKS" default:"{}"`
	//SocialMediaLinks map[string]string `json:"social_media_links" env:"SOCIAL_MEDIA_LINKS" default:"{}"`
}

// SingleLineAddress returns a single line address string
func (c CompanyConfig) SingleLineAddress() string {
	if c.Address2 != "" {
		return fmt.Sprintf("%s %s, %s, %s %s", c.Address, c.Address2, c.City, c.State, c.Zip)
	}
	return fmt.Sprintf("%s, %s, %s %s", c.Address, c.City, c.State, c.Zip)
}

// TwoLineAddress returns two lines of address strings (the first line contains the address and the second line contains the city, state, and ZIP code)
func (c CompanyConfig) TwoLineAddress() (string, string) {
	line1 := c.Address
	if c.Address2 != "" {
		line1 = fmt.Sprintf("%s, %s", c.Address, c.Address2)
	}
	line2 := fmt.Sprintf("%s, %s %s", c.City, c.State, c.Zip)
	return line1, line2
}

type LogConfig struct {
	Format      string `json:"format" env:"LOG_FORMAT" default:"pretty"`
	IncludeTime bool   `json:"include_time" env:"LOG_INCLUDE_TIME" default:"false"`
	Level       string `json:"level" env:"LOG_LEVEL" default:"debug"`
	Verbose     bool   `json:"verbose" env:"LOG_VERBOSE" default:"false"`
}
