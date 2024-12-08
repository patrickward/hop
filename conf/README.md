# Configuration Package

The `conf` package provides a flexible, layered configuration system for Go applications with support for JSON files, environment variables, and validation. It allows for hierarchical configuration with multiple sources and automatic environment variable mapping.

## Features

- JSON configuration file support with hierarchical override system
- Automatic environment variable mapping
- Configuration file discovery based on environment
- Default value support via struct tags
- Validation system
- Secret value masking in logs/output
- Duration string parsing
- Pretty printing of configuration values

## Basic Usage

```go
type MyConfig struct {
    Hop conf.HopConfig        `json:"hop"`
    Database struct {
        Host     string       `json:"host" default:"localhost"`
        Port     int          `json:"port" default:"5432"`
        Password string       `json:"password" secret:"true"`
    } `json:"database"`
}

func main() {
    config := &MyConfig{}
    manager := conf.NewManager(config,
        conf.WithEnvironment("development"),
        conf.WithEnvPrefix("APP"),
    )
    
    if err := manager.Load(); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration Sources

The configuration system loads values in the following order (later sources override earlier ones):

1. Default values from struct tags
2. Discovered configuration files
3. Explicitly specified configuration files
4. Environment variables

## Configuration File Discovery

The system automatically discovers and loads configuration files in the following order:

```
config.json              # Base configuration
config.local.json       # Local overrides
config/config.json      # Config directory
config/config.local.json # Config directory local overrides
config.<env>.json       # Environment-specific config
config/<env>.json       # Environment-specific in directory
config/config.<env>.json # Environment-specific in config directory
```

## Manager Options

The configuration manager supports several options for customization:

```go
manager := conf.NewManager(config,
    // Set environment for file discovery
    conf.WithEnvironment("development"),
    
    // Add environment variable prefix
    conf.WithEnvPrefix("APP"),
    
    // Add specific config files
    conf.WithConfigFile("config/custom.json"),
    
    // Add multiple config files
    conf.WithConfigFiles("config/base.json", "config/override.json"),
    
    // Add all JSON files from a directory
    conf.WithDefaultConfigDir("config"),
)
```

## Struct Tags

The package supports several struct tags for configuration:

- `json`: Specifies the JSON field name
- `default`: Sets the default value
- `secret`: Marks sensitive values for masking in output

Example:
```go
type Config struct {
    Port     int    `json:"port" default:"8080"`
    Host     string `json:"host" default:"localhost"`
    APIKey   string `json:"api_key" secret:"true"`
    Timeout  conf.Duration `json:"timeout" default:"5m"`
}
```

## Environment Variables

Environment variables are automatically mapped to configuration fields using the following rules:

1. Field names are converted to SCREAMING_SNAKE_CASE
2. Nested structs use underscore separation
3. Optional prefix is prepended if specified

Example mapping:
```
Config struct:
    Database.Host -> DATABASE_HOST
    Database.MaxConnections -> DATABASE_MAX_CONNECTIONS

With prefix "APP":
    Database.Host -> APP_DATABASE_HOST
    Database.MaxConnections -> APP_DATABASE_MAX_CONNECTIONS
```

## Duration Support

The package includes a special `Duration` type that supports parsing duration strings in both JSON and environment variables:

```go
type ServerConfig struct {
    ReadTimeout  conf.Duration `json:"read_timeout" default:"15s"`
    WriteTimeout conf.Duration `json:"write_timeout" default:"15s"`
}
```

Supported duration formats: "300ms", "1.5h", "2h45m", etc.

## Validation

The configuration system supports two types of validation:

1. Framework validation (ensuring required Hop framework configuration)
2. Custom validation via the `Validator` interface

To implement custom validation:

```go
func (c *MyConfig) Validate() error {
    if c.Database.Port < 1024 {
        return fmt.Errorf("database port must be > 1024")
    }
    return nil
}
```

## Pretty Printing

The configuration manager includes pretty printing support with automatic masking of sensitive values:

```go
fmt.Println(manager.String())

// Output:
// Database.Host                              = "localhost"
// Database.Port                              = 5432
// Database.Password                          = [REDACTED] "p***d"
```

## Reloading Configuration

The configuration can be reloaded at runtime:

```go
if err := manager.Reload(); err != nil {
    log.Printf("Failed to reload configuration: %v", err)
}
```

## Thread Safety

The configuration manager is thread-safe and can be safely accessed from multiple goroutines. All read and write operations are protected by appropriate mutex locks.

## Error Handling

The configuration system provides detailed error messages for:

- File loading failures
- Environment variable parsing errors
- Validation failures
- Type conversion errors
- Missing required configurations

## Best Practices

1. Always embed the `conf.HopConfig` struct in your configuration
2. Use environment-specific files for different deployment environments
3. Keep sensitive values in environment variables rather than config files
4. Use the `secret` tag for sensitive values to prevent logging exposure
5. Implement the `Validator` interface for custom validation rules
6. Use strongly-typed configuration structs rather than maps
7. Provide sensible defaults using the `default` tag

## Example Complete Configuration

```go
type Config struct {
    Hop conf.HopConfig `json:"hop"`
    
    Database struct {
        Host            string        `json:"host" default:"localhost"`
        Port            int           `json:"port" default:"5432"`
        User            string        `json:"user" default:"postgres"`
        Password        string        `json:"password" secret:"true"`
        MaxConnections  int           `json:"max_connections" default:"100"`
        ConnTimeout     conf.Duration `json:"conn_timeout" default:"10s"`
    } `json:"database"`
    
    Redis struct {
        Host     string        `json:"host" default:"localhost"`
        Port     int           `json:"port" default:"6379"`
        Timeout  conf.Duration `json:"timeout" default:"5s"`
    } `json:"redis"`
    
    API struct {
        Endpoint    string        `json:"endpoint" default:"http://api.example.com"`
        Timeout     conf.Duration `json:"timeout" default:"30s"`
        MaxRetries  int          `json:"max_retries" default:"3"`
        APIKey      string       `json:"api_key" secret:"true"`
    } `json:"api"`
}

func (c *Config) Validate() error {
    if c.Database.MaxConnections < 1 {
        return fmt.Errorf("database max connections must be positive")
    }
    if c.API.MaxRetries < 0 {
        return fmt.Errorf("api max retries cannot be negative")
    }
    return nil
}
```
