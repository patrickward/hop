package conf

import (
	"fmt"
	"reflect"
)

// Validator interface allows configs to implement their own validation
type Validator interface {
	Validate() error
}

// HopConfigValidator implements core framework validation
type HopConfigValidator struct{}

func (v *HopConfigValidator) Validate(cfg interface{}) error {
	// First check if config has framework configuration
	if err := v.validateHopConfig(cfg); err != nil {
		return fmt.Errorf("hop framework validation failed: %w", err)
	}

	// Then check if the config implements Validator interface
	if validator, ok := cfg.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("application validation failed: %w", err)
		}
	}

	return nil
}

// validateHopConfig ensures required Hop framework config is present
func (v *HopConfigValidator) validateHopConfig(cfg interface{}) error {
	val := reflect.ValueOf(cfg)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("configuration must be a struct")
	}

	hopField := val.FieldByName("Hop")
	if !hopField.IsValid() {
		return fmt.Errorf("configuration must include Hop framework configuration as 'Hop' field")
	}

	if hopField.Type() != reflect.TypeOf(HopConfig{}) {
		return fmt.Errorf("Hop field must be of type HopConfig")
	}

	return nil
}
