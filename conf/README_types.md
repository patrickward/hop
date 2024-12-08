# Configuration Types Guide

## Built-in Types

The configuration system includes several built-in types in the `conftype` package for common configuration needs:

### Duration
```go
import "github.com/patrickward/hop/conf/conftype"

type Config struct {
    Timeout conftype.Duration `json:"timeout" default:"30s"`
}
```
Parses duration strings like "300ms", "1.5h", "2h45m". Uses Go's standard duration format.

### StringList
```go
type Config struct {
    Tags conftype.StringList `json:"tags" default:"web,api,prod"`
}
```
Handles comma-separated string lists. Can be specified as JSON arrays or comma-separated strings.

### LogLevel
```go
type Config struct {
    Level conftype.LogLevel `json:"level" default:"info"`
}
```
Validates against standard log levels: debug, info, warn, error.

### ByteSize
```go
type Config struct {
    MaxMemory conftype.ByteSize `json:"max_memory" default:"512MB"`
}
```
Parses human-readable byte sizes like "512MB", "1.5GB", "2TB". Supports B, KB, MB, GB, TB, PB suffixes.

### TCPAddress
```go
type Config struct {
    Server conftype.TCPAddress `json:"server" default:"localhost:8080"`
}
```
Validates and parses host:port combinations with proper port range checking.

### TimeZone
```go
type Config struct {
    Location conftype.TimeZone `json:"location" default:"America/New_York"`
}
```
Validates against valid time zone names and provides a string representation.

## Basic JSON Types

The configuration system supports all standard JSON types:

```go
type Config struct {
    // String
    Name string `json:"name" default:"default-name"`
    
    // Boolean
    Enabled bool `json:"enabled" default:"true"`
    
    // Integer types
    Port int `json:"port" default:"8080"`
    SmallValue int8 `json:"small_value"`
    MediumValue int16 `json:"medium_value"`
    LargeValue int32 `json:"large_value"`
    HugeValue int64 `json:"huge_value"`
    
    // Unsigned integer types
    UnsignedPort uint `json:"unsigned_port"`
    SmallUnsigned uint8 `json:"small_unsigned"`
    MediumUnsigned uint16 `json:"medium_unsigned"`
    LargeUnsigned uint32 `json:"large_unsigned"`
    HugeUnsigned uint64 `json:"huge_unsigned"`
    
    // Float types
    Factor float32 `json:"factor"`
    Precise float64 `json:"precise"`
}
```

## Adding Custom Types

You can add your own custom types by implementing three interfaces:

1. `StringParser` for environment variables and defaults
2. `json.Marshaler` and `json.Unmarshaler` for JSON support
3. `fmt.Stringer` for string representation (recommended)

Here are some examples of useful custom types you might want to implement:

### Email Address
```go
type EmailAddress string

func (e *EmailAddress) ParseString(s string) error {
    // Validate email format
    if !strings.Contains(s, "@") {
        return fmt.Errorf("invalid email address: %s", s)
    }
    *e = EmailAddress(s)
    return nil
}

func (e EmailAddress) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(e))
}

func (e *EmailAddress) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    return e.ParseString(s)
}

func (e EmailAddress) String() string {
    return string(e)
}
```

### RegexPattern
```go
type RegexPattern struct {
    pattern *regexp.Regexp
}

func (r *RegexPattern) ParseString(s string) error {
    pattern, err := regexp.Compile(s)
    if err != nil {
        return fmt.Errorf("invalid regex pattern: %w", err)
    }
    r.pattern = pattern
    return nil
}

func (r RegexPattern) MarshalJSON() ([]byte, error) {
    if r.pattern == nil {
        return []byte(`""`), nil
    }
    return json.Marshal(r.pattern.String())
}

func (r *RegexPattern) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    return r.ParseString(s)
}

func (r RegexPattern) String() string {
    if r.pattern == nil {
        return ""
    }
    return r.pattern.String()
}
```

Other useful types you might consider implementing:
- FilePath (with existence/permission validation)
- DatabaseURL (with connection string parsing)
- JSONData (for embedded JSON objects)
- TimeZone (with validation against valid zones)
- CIDRRange (for network subnet configuration)
- CSVData (for structured comma-separated data)

## Best Practices

1. Always implement both JSON and string parsing
2. Provide clear error messages for invalid values
3. Handle empty/zero values gracefully
4. Include proper validation
5. Use appropriate types for the data (don't just use string for everything)
6. Document any specific format requirements
7. Add thorough tests for all parsing scenarios

## Testing Your Types

Be sure to test:
1. JSON marshaling and unmarshaling
2. String parsing for environment variables
3. Default value handling
4. Error cases
5. Edge cases (empty values, invalid formats)
6. String representation

Example:
```go
func TestEmailAddress(t *testing.T) {
    cases := []struct {
        input    string
        valid    bool
        expected string
    }{
        {"user@example.com", true, "user@example.com"},
        {"invalid", false, ""},
        {"", true, ""},
    }

    for _, tc := range cases {
        t.Run(tc.input, func(t *testing.T) {
            var email EmailAddress
            err := email.ParseString(tc.input)
            if tc.valid {
                assert.NoError(t, err)
                assert.Equal(t, tc.expected, email.String())
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```
