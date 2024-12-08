package conf

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/patrickward/hop/conf/conftype"
)

// EnvParser handles environment variable parsing for configuration structs
type EnvParser struct {
	// namespace is an optional prefix for all environment variables
	namespace string
}

// NewEnvParser creates a new environment variable parser
func NewEnvParser(namespace string) *EnvParser {
	return &EnvParser{
		namespace: strings.TrimRight(strings.ToUpper(namespace), "_"),
	}
}

// Parse walks through the given struct and populates it from environment variables
func (p *EnvParser) Parse(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("value must be a pointer to struct")
	}
	return p.ParseStruct(val.Elem(), "")
}

// ParseStruct handles parsing for struct values
func (p *EnvParser) ParseStruct(val reflect.Value, prefix string) error {
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("value must be a struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		// Convert the field name to environment variable format
		envFieldName := ToScreamingSnake(structField.Name)

		// Build the full path for this field
		var envPath string
		if prefix == "" {
			envPath = envFieldName
		} else {
			envPath = prefix + "_" + envFieldName
		}

		// Handle nested structs (except Duration which is a special case)
		if field.Kind() == reflect.Struct && structField.Type != reflect.TypeOf(conftype.Duration{}) {
			if err := p.ParseStruct(field, envPath); err != nil {
				return fmt.Errorf("parsing nested struct %s: %w", structField.Name, err)
			}
			continue
		}

		// Add namespace prefix if it exists
		fullEnvName := envPath
		if p.namespace != "" {
			fullEnvName = p.namespace + "_" + envPath
		}

		// Look for environment variable
		if value, exists := os.LookupEnv(fullEnvName); exists {
			if err := setFieldValue(field, value); err != nil {
				return fmt.Errorf("setting field %s from env %s: %w", structField.Name, fullEnvName, err)
			}
		}
	}

	return nil
}

// ParseStructOLD handles parsing for struct values
func (p *EnvParser) ParseStructOLD(val reflect.Value, prefix string) error {
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("value must be a struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		structField := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		// Build the environment variable name from the path
		envName := p.buildEnvName(prefix, structField.Name)
		fmt.Println("envName: ", envName)

		// Handle nested structs (except Duration which is a special case)
		if field.Kind() == reflect.Struct && structField.Type != reflect.TypeOf(conftype.Duration{}) {
			if err := p.ParseStruct(field, envName); err != nil {
				return fmt.Errorf("parsing nested struct %s: %w", structField.Name, err)
			}
			continue
		}

		// Look for environment variable
		if value, exists := os.LookupEnv(envName); exists {
			if err := setFieldValue(field, value); err != nil {
				return fmt.Errorf("setting field %s from env %s: %w", structField.Name, envName, err)
			}
		}
	}

	return nil
}

// buildEnvName creates the environment variable name from the path
func (p *EnvParser) buildEnvName(prefix, fieldName string) string {
	// Convert field name to SCREAMING_SNAKE_CASE
	name := ToScreamingSnake(fieldName)

	// Build full name with prefix
	var parts []string
	if p.namespace != "" {
		parts = append(parts, strings.ToUpper(p.namespace))
	}
	if prefix != "" {
		parts = append(parts, prefix)
	}
	parts = append(parts, name)

	return strings.Join(parts, "_")
}

// ToScreamingSnake converts a string from camelCase/PascalCase to SCREAMING_SNAKE_CASE
func ToScreamingSnake(s string) string {
	var result strings.Builder
	result.Grow(len(s) + 3)

	for i, r := range s {
		// Add underscore if:
		// 1. Not the first character AND
		// 2. Current character is uppercase AND
		// 3. Either:
		//    - Next character is lowercase OR
		//    - Previous character is lowercase
		if i > 0 && isUpper(r) {
			if (i+1 < len(s) && isLower(rune(s[i+1]))) ||
				(i-1 >= 0 && isLower(rune(s[i-1]))) {
				result.WriteRune('_')
			}
		}
		result.WriteRune(toUpper(r))
	}

	return result.String()
}

// Helper functions for character case conversion
//func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
//func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
//func toUpper(r rune) bool { return r - ('a' - 'A') }

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - ('a' - 'A')
	}
	return r
}
