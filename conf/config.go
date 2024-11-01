// Package conf provides a way to load configuration from JSON files and environment variables,
// along with a structure to hold the configuration settings for an application and the ability
// to set up command-line flags for configuration options.
package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

// BaseConfig provides core configuration options
type BaseConfig struct {
	//Environment string            `json:"environment" env:"APP_ENVIRONMENT" default:"development"`
	//Debug       bool              `json:"debug" env:"APP_DEBUG" default:"false"`
	App         AppConfig         `json:"app"`
	Company     CompanyConfig     `json:"company"`
	Maintenance MaintenanceConfig `json:"maintenance"`
	Server      ServerConfig      `json:"server"`
	Database    DatabaseConfig    `json:"database"`
	SMTP        SMPTConfig        `json:"smtp"`
	Log         LogConfig         `json:"log"`
}

func (c *BaseConfig) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *BaseConfig) IsProduction() bool {
	return c.App.Environment == "production"
}

type AppConfig struct {
	Environment string `json:"environment" env:"APP_ENVIRONMENT" default:"development"`
	Debug       bool   `json:"debug" env:"APP_DEBUG" default:"false"`
	SAASMode    bool   `json:"saas_mode" env:"SAAS_MODE" default:"false"`
}

type MaintenanceConfig struct {
	Enabled bool   `json:"enabled" env:"MAINTENANCE_ENABLED" default:"false"`
	Message string `json:"message" env:"MAINTENANCE_MESSAGE" default:""`
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
	Name         string `json:"name" env:"COMPANY_NAME" default:""`
	LogoURL      string `json:"logo_url" env:"COMPANY_LOGO_URL" default:""`
	SupportEmail string `json:"support_email" env:"COMPANY_SUPPORT_EMAIL" default:""`
	WebsiteName  string `json:"website_name" env:"COMPANY_WEBSITE_NAME" default:""`
	WebsiteURL   string `json:"website_url" env:"COMPANY_WEBSITE_URL" default:""`
	//SiteLinks        map[string]string `json:"site_links" env:"SITE_LINKS" default:"{}"`
	//SocialMediaLinks map[string]string `json:"social_media_links" env:"SOCIAL_MEDIA_LINKS" default:"{}"`
}

type LogConfig struct {
	Format      string `json:"format" env:"LOG_FORMAT" default:"pretty"`
	IncludeTime bool   `json:"include_time" env:"LOG_INCLUDE_TIME" default:"false"`
	Level       string `json:"level" env:"LOG_LEVEL" default:"debug"`
	Verbose     bool   `json:"verbose" env:"LOG_VERBOSE" default:"false"`
}

// Load sets the defaults and loads configuration from JSON files and environment variables
func Load(cfg interface{}, files ...string) error {
	// Set defaults first
	if err := setDefaults(cfg); err != nil {
		return fmt.Errorf("setting defaults: %w", err)
	}

	// Load JSON files in order
	for _, file := range files {
		if err := loadFile(cfg, file); err != nil {
			return fmt.Errorf("loading config file %s: %w", file, err)
		}
	}

	// Override with environment variables
	if err := loadEnv(cfg); err != nil {
		return fmt.Errorf("loading environment variables: %w", err)
	}

	return nil
}

// setDefaults sets default values based on struct tags
func setDefaults(cfg interface{}) error {
	val := reflect.ValueOf(cfg)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer")
	}
	return setDefaultsStruct(val.Elem())
}

// setDefaultsStruct recursively sets default values for struct fields
func setDefaultsStruct(val reflect.Value) error {
	if val.Kind() != reflect.Struct {
		return nil // Skip non-structs, they'll be handled by their parent's field iteration
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Handle embedded structs
		if typeField.Anonymous {
			if err := setDefaultsStruct(field); err != nil {
				return err
			}
			continue
		}

		// Get default value from tag
		defaultVal := typeField.Tag.Get("default")
		if defaultVal != "" {
			if err := setFieldValue(field, defaultVal); err != nil {
				return fmt.Errorf("setting default for field %s: %w", typeField.Name, err)
			}
		}

		// If it's a struct, recurse into it
		// Special case for Duration which is a struct but should be treated as a value
		if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(Duration{}) {
			if err := setDefaultsStruct(field); err != nil {
				return err
			}
		}
	}

	return nil
}

// loadFile loads configuration from a JSON file
func loadFile(cfg interface{}, file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Skip missing files
		}
		return err
	}

	return json.Unmarshal(data, cfg)
}

// loadEnv loads configuration from environment variables
func loadEnv(cfg interface{}) error {
	val := reflect.ValueOf(cfg)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a pointer")
	}
	return loadEnvStruct(val.Elem(), "")
}

// loadEnvStruct recursively loads environment variables into struct fields
func loadEnvStruct(val reflect.Value, prefix string) error {
	if val.Kind() != reflect.Struct {
		return nil // Skip non-structs, they'll be handled by their parent's field iteration
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// Handle embedded structs
		if typeField.Anonymous {
			if err := loadEnvStruct(field, prefix); err != nil {
				return err
			}
			continue
		}

		// Get environment variable name from tag
		envName := typeField.Tag.Get("env")
		if envName == "" {
			// If no env tag but it's a struct, recurse into it
			// Skip Duration type as it's handled as a value
			if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(Duration{}) {
				newPrefix := prefix
				if prefix != "" {
					newPrefix = prefix + "_" + typeField.Name
				}
				if err := loadEnvStruct(field, newPrefix); err != nil {
					return err
				}
			}
			continue
		}

		// Check if environment variable exists
		if value, exists := os.LookupEnv(envName); exists {
			if err := setFieldValue(field, value); err != nil {
				return fmt.Errorf("setting field %s: %w", typeField.Name, err)
			}
		}
	}

	return nil
}

// setFieldValue sets the value of a struct field based on its type
func setFieldValue(field reflect.Value, value string) error {
	// Ensure the field is settable
	if !field.CanSet() {
		return nil
	}

	switch field.Type().String() {
	case "conf.Duration":
		d, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(Duration{d}))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type())
	}
	return nil
}
