# Adding Custom Types to the Configuration System

## Overview

The configuration system supports custom types through Go's type system and interfaces. To add a custom type, you need to:

1. Implement `json.Unmarshaler` for JSON file support
2. Handle type conversion from string values for environment variables

Our `Duration` type serves as a good example of how to implement a custom type:

```go
// Duration wraps time.Duration for custom parsing
type Duration struct {
    time.Duration
}

// UnmarshalJSON implements json.Unmarshaler for JSON file support
func (d *Duration) UnmarshalJSON(b []byte) error {
    var v interface{}
    if err := json.Unmarshal(b, &v); err != nil {
        return err
    }
    
    switch value := v.(type) {
    case float64:
        d.Duration = time.Duration(value)
        return nil
    case string:
        var err error
        d.Duration, err = time.ParseDuration(value)
        if err != nil {
            return err
        }
        return nil
    default:
        return fmt.Errorf("invalid duration")
    }
}

// String returns the string representation of the duration
func (d Duration) String() string {
    return d.Duration.String()
}
```

## Environment Variable Support

For environment variables to work with your custom type, you need to add support in the `setFieldValue` function:

```go
func setFieldValue(field reflect.Value, value string) error {
    switch field.Type().String() {
    case "conf.Duration":
        d, err := time.ParseDuration(value)
        if err != nil {
            return err
        }
        field.Set(reflect.ValueOf(Duration{d}))
        return nil
    // ... other cases
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
    // ... other cases
    }
    return nil
}
```

## Example: Adding a Custom URL Type

Here's how to add a new URL type to the configuration system:

```go
// URL wraps url.URL for configuration use
type URL struct {
    *url.URL
}

// UnmarshalJSON implements json.Unmarshaler
func (u *URL) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil {
        return err
    }
    
    parsed, err := url.Parse(s)
    if err != nil {
        return err
    }
    
    u.URL = parsed
    return nil
}

// String implements Stringer
func (u URL) String() string {
    if u.URL == nil {
        return ""
    }
    return u.URL.String()
}
```

Then add support in `setFieldValue`:

```go
func setFieldValue(field reflect.Value, value string) error {
    switch field.Type().String() {
    case "conf.URL":
        parsed, err := url.Parse(value)
        if err != nil {
            return err
        }
        field.Set(reflect.ValueOf(URL{parsed}))
        return nil
    // ... other cases
    }
    // ... rest of implementation
}
```

## Usage

Use your custom type in configuration structs:

```go
type Config struct {
    Server struct {
        BaseURL URL      `json:"base_url" default:"http://localhost:8080"`
        Timeout Duration `json:"timeout" default:"30s"`
    } `json:"server"`
}
```

## Testing Custom Types

Test both JSON unmarshaling and environment variable parsing:

```go
func TestURLParsing(t *testing.T) {
    // Test JSON unmarshaling
    var u URL
    err := json.Unmarshal([]byte(`"http://example.com"`), &u)
    require.NoError(t, err)
    assert.Equal(t, "http://example.com", u.String())

    // Test environment variable parsing
    cfg := &Config{}
    mgr := conf.NewManager(cfg)
    
    os.Setenv("SERVER_BASE_URL", "http://example.com")
    err = mgr.Load()
    require.NoError(t, err)
    assert.Equal(t, "http://example.com", cfg.Server.BaseURL.String())
}
```

## Best Practices

1. Always implement `String()` for pretty printing support
2. Provide clear error messages for parsing failures
3. Handle empty/zero values gracefully
4. Test both JSON and environment variable parsing
5. Document any specific formatting requirements

## Implementation Notes

- The system uses Go's reflection to handle type conversion
- Custom types need to be handled explicitly in `setFieldValue`
- JSON unmarshaling is handled through the standard `encoding/json` package
- Default values are parsed using the same mechanism as environment variables

## Limitations

- Custom types must be added to `setFieldValue` explicitly
- No support for generic type conversion
- Default values must be parseable as strings
