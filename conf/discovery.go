package conf

import "fmt"

// ConfigDiscovery handles automatic configuration file discovery
type configDiscovery struct {
	environment string
}

// defaultPaths returns the base configuration paths that should always be checked
func (d *configDiscovery) defaultPaths() []string {
	return []string{
		"config.json",              // base config
		"config.local.json",        // local overrides
		"config/config.json",       // config directory
		"config/config.local.json", // config directory local overrides
	}
}

// environmentPaths returns environment-specific paths if environment is set
func (d *configDiscovery) environmentPaths() []string {
	if d.environment == "" {
		return nil
	}

	return []string{
		fmt.Sprintf("config.%s.json", d.environment),
		fmt.Sprintf("config/%s.json", d.environment),
		fmt.Sprintf("config/config.%s.json", d.environment),
	}
}

// paths returns all potential configuration file paths in load order
func (d *configDiscovery) paths() []string {
	// Start with default paths
	paths := d.defaultPaths()

	// Add environment-specific paths if environment is set
	if d.environment != "" {
		paths = append(paths, d.environmentPaths()...)
	}

	return paths
}
