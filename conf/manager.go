package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Manager handles configuration loading and access
type Manager struct {
	mu        sync.RWMutex
	config    interface{}
	files     []string
	envParser *EnvParser
	validator *HopConfigValidator
	discovery *configDiscovery
}

// Option is a functional option for Manager
type Option func(*Manager)

// NewManager creates a Manager instance
func NewManager(config interface{}, opts ...Option) *Manager {
	m := &Manager{
		config:    config,
		envParser: NewEnvParser(""),
		validator: &HopConfigValidator{},
		discovery: &configDiscovery{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// WithEnvPrefix sets the environment variable prefix, which makes the env parser look for variables with the prefix.
// For example, if the prefix is "APP", the parser will only look for environment variables like "APP_PORT" and "APP_DEBUG".
func WithEnvPrefix(prefix string) Option {
	return func(m *Manager) {
		m.envParser = NewEnvParser(prefix)
	}
}

// WithConfigFile adds a JSON file to the list of configuration files to load
// Files are processed in the order they are added
func WithConfigFile(file string) Option {
	return func(m *Manager) {
		m.files = append(m.files, file)
	}
}

// WithConfigFiles adds multiple JSON files to the list of configuration files to load
// Files are processed in the order they are added
func WithConfigFiles(files ...string) Option {
	return func(m *Manager) {
		m.files = append(m.files, files...)
	}
}

// WithDefaultConfigDir adds all .json files from a directory to the list of configuration files to load
func WithDefaultConfigDir(dir string) Option {
	return func(m *Manager) {
		// list all .json files in the directory
		files, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			return
		}
		m.files = append(m.files, files...)
	}
}

// WithEnvironment sets the environment for configuration file discovery
func WithEnvironment(env string) Option {
	return func(m *Manager) {
		m.discovery = &configDiscovery{
			environment: strings.ToLower(env),
		}
	}
}

// doLoad initializes the configuration in a specific order:
// 1. Set defaults from struct tags
// 2. Load JSON files in order specified
// 3. Override with environment variables
func (m *Manager) doLoad(cfg interface{}) error {
	// Set defaults first
	if err := m.setDefaults(cfg); err != nil {
		return fmt.Errorf("error setting defaults: %w", err)
	}

	//// If no files specified, check for default config.json
	//if len(m.files) == 0 {
	//	if _, err := os.Stat("config.json"); err == nil {
	//		m.files = append(m.files, "config.json")
	//	}
	//}

	// Load discovered files first
	if m.discovery != nil {
		for _, path := range m.discovery.paths() {
			if err := m.loadFile(path); err != nil {
				return fmt.Errorf("error loading file %s: %w", path, err)
			}
		}
	}

	// Load JSON files in order
	for _, file := range m.files {
		if err := m.loadFile(file); err != nil {
			return fmt.Errorf("error loading file %s: %w", file, err)
		}
	}

	// Override with environment variables
	if err := m.envParser.Parse(cfg); err != nil {
		return fmt.Errorf("error parsing environment variables: %w", err)
	}

	// Run validation after all loading is complete
	if err := m.validator.Validate(cfg); err != nil {
		return fmt.Errorf("error validating config: %w", err)
	}

	return nil
}

// Load performs initial load with lock
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.doLoad(m.config)
}

// Reload safely reloads config with new values
func (m *Manager) Reload() error {
	newCfg := reflect.New(reflect.TypeOf(m.config).Elem()).Interface()

	if err := m.doLoad(newCfg); err != nil {
		return err
	}

	m.mu.Lock()
	// Copy values to existing config
	reflect.ValueOf(m.config).Elem().Set(reflect.ValueOf(newCfg).Elem())
	m.mu.Unlock()

	return nil
}

// Get returns the current configuration
func (m *Manager) Get() interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// setDefaults sets default values for the configuration struct
func (m *Manager) setDefaults(cfg interface{}) error {
	return setDefaultsStruct(reflect.ValueOf(cfg).Elem())
}

// loadFile loads a single JSON file into the configuration struct
func (m *Manager) loadFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Skip missing files
		}
		return err
	}

	return json.Unmarshal(data, m.config)
}

// Helper functions

// setDefaultsStruct sets default values for a struct
func setDefaultsStruct(val reflect.Value) error {
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		// Handle embedded structs
		if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(Duration{}) {
			if err := setDefaultsStruct(field); err != nil {
				return err
			}
			continue
		}

		// Set default if tag exists
		if defaultVal := typeField.Tag.Get("default"); defaultVal != "" {
			if err := setFieldValue(field, defaultVal); err != nil {
				return fmt.Errorf("field %s: %w", typeField.Name, err)
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

// String returns a pretty string representation of the configuration
func (m *Manager) String() string {
	return PrettyString(m.config)
}
