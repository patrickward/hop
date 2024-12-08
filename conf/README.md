# Hop Configuration System

The Hop configuration system provides a flexible, type-safe way to manage application configuration through multiple sources including environment variables, JSON files, and default values.

## Core Features

- Multiple configuration sources with clear precedence
- Type-safe configuration through Go structs
- Environment variable support with optional prefixing
- Default values through struct tags
- Validation support (both framework and application-specific)
- Configuration reloading
- Pretty printing with sensitive value masking

## Basic Usage

### 1. Define Your Configuration

Create a struct that includes the Hop framework configuration and your application-specific settings:

```go
type AppConfig struct {
    // Required: Hop framework configuration
    Hop conf.HopConfig `json:"hop"`

    // Application-specific configuration
    API struct {
        Endpoint    string        `json:"endpoint" default:"http://api.local"`
        Timeout     conf.Duration `json:"timeout" default:"30s"`
        MaxRetries  int          `json:"max_retries" default:"3"`
        SecretKey   string       `json:"secret_key"`
    } `json:"api"`
}
```

### 2. Initialize and Load

```go
// Create config instance
cfg := &AppConfig{}

// Create manager with options
mgr := conf.NewManager(cfg,
    conf.WithConfigFile("config.json"),
    conf.WithEnvPrefix("APP")
)

// Load configuration
if err := mgr.Load(); err != nil {
    log.Fatal(err)
}
```

## Configuration Sources and Precedence

The system loads configuration in the following order (later sources override earlier ones):

1. Default values (from struct tags)
2. JSON configuration files (in order specified)
3. Environment variables

## Environment Variables

Environment variables are automatically mapped from struct fields using SCREAMING_SNAKE_CASE:

- Struct field: `API.Endpoint` → Env var: `APP_API_ENDPOINT` (with prefix)
- Struct field: `API.MaxRetries` → Env var: `APP_API_MAX_RETRIES` (with prefix)

## Default Values

Set defaults using the `default` struct tag:

```go
type Config struct {
    Server struct {
        Port int `json:"port" default:"8080"`
    } `json:"server"`
}
```

## Configuration Files

JSON configuration files can be specified using the `WithConfigFile` or `WithConfigFiles` options:

```go
mgr := conf.NewManager(cfg,
    conf.WithConfigFiles("config.json", "config.local.json"),
)
```

## Validation

### Framework Validation

The system automatically validates that required Hop framework configuration is present.

### Custom Validation

Implement the `Validator` interface to add application-specific validation:

```go
func (c *AppConfig) Validate() error {
    if c.API.MaxRetries < 0 {
        return fmt.Errorf("max retries cannot be negative")
    }
    return nil
}
```

## Configuration Reloading

Reload configuration at runtime:

```go
if err := mgr.Reload(); err != nil {
    log.Printf("Failed to reload configuration: %v", err)
}
```

## Pretty Printing

Print the current configuration with sensitive value masking:

```go
// Via manager
fmt.Println(mgr.String())

// Or directly
fmt.Println(conf.PrettyString(cfg))
```

Output example:
```
Server.Host                                 = "localhost"
Server.Port                                = 8080
API.Endpoint                               = "http://api.example.com"
API.SecretKey                              = "a***t"
API.MaxRetries                             = 3
```

## Duration Type

The system includes a custom `Duration` type for parsing time durations:

```go
type Config struct {
    Timeout conf.Duration `json:"timeout" default:"30s"`
}
```

Supported formats: "300ms", "1.5h", "2h45m", etc.

## Advanced Features

### Environment Prefixing

Add a prefix to all environment variables:

```go
mgr := conf.NewManager(cfg, conf.WithEnvPrefix("MYAPP"))
```

### Sensitive Value Masking

The pretty printer automatically masks values for fields containing these patterns:
- password
- secret
- key
- token
- credential

## Best Practices

1. Always embed the Hop framework configuration:
   ```go
   type AppConfig struct {
       Hop conf.HopConfig `json:"hop"`
       // ... app config
   }
   ```

2. Use meaningful default values:
   ```go
   `default:"5s"` // Clear duration
   `default:"100"` // Reasonable limit
   ```

3. Implement custom validation for critical settings:
   ```go
   func (c *AppConfig) Validate() error {
       // Validate configuration
   }
   ```

4. Use environment variables for secrets:
   ```go
   SecretKey string `json:"secret_key"` // Set via APP_SECRET_KEY
   ```

## Error Handling

The system provides detailed error messages for:
- Invalid configuration files
- Invalid environment variables
- Validation failures
- Type conversion errors

Example:
```go
if err := mgr.Load(); err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

## Limitations

- JSON files only (no YAML or TOML support)
- No dynamic configuration updates (must call Reload)
- No configuration history tracking
- No configuration schema versioning

## Contributing

Before adding new features, consider:
1. Backward compatibility
2. Error handling
3. Documentation updates
4. Test coverage

