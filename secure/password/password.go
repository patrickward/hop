// Package password provides functions for hashing and verifying passwords.
//
// Example usage:
// hashedPassword, err := password.Hash("mysecretpassword")
// isValid, err := password.Matches("mysecretpassword", hashedPassword)
package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrEmptyPassword indicates that an empty password was provided
	ErrEmptyPassword = errors.New("password cannot be empty")
	// ErrEmptyHash indicates that an empty hash was provided for verification
	ErrEmptyHash = errors.New("hash cannot be empty")
)

const (
	// MinCost is the minimum bcrypt cost factor
	MinCost = bcrypt.MinCost // 4
	// MaxCost is the maximum bcrypt cost factor
	MaxCost = bcrypt.MaxCost // 31
	// DefaultCost is the default cost factor used if not specified
	// or if specified cost is invalid
	DefaultCost = 12
	// RecommendedMinCost is the recommended minimum cost for production use
	RecommendedMinCost = 10
)

// config holds the password hashing configuration
type config struct {
	cost int
}

// Option defines a function that can modify the config
type Option func(*config)

// WithCost sets the bcrypt cost parameter
// If cost is outside valid range (MinCost to MaxCost),
// DefaultCost will be used instead
func WithCost(cost int) Option {
	return func(c *config) {
		if cost >= MinCost && cost <= MaxCost {
			c.cost = cost
		} else {
			c.cost = DefaultCost
		}
	}
}

// defaultConfig returns the default configuration
func defaultConfig() *config {
	return &config{
		cost: DefaultCost,
	}
}

// validateConfig ensures the configuration is valid
func validateConfig(cfg *config) {
	// Ensure cost is within valid range
	if cfg.cost < MinCost || cfg.cost > MaxCost {
		cfg.cost = DefaultCost
	}
}

// Hash creates a password hash with the given options
// If no options are provided, DefaultCost will be used
// Returns ErrEmptyPassword if the password is empty
func Hash(plaintext string, opts ...Option) (string, error) {
	if plaintext == "" {
		return "", ErrEmptyPassword
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	validateConfig(cfg)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintext), cfg.cost)
	if err != nil {
		return "", fmt.Errorf("password hash: %w", err)
	}

	return string(hashedPassword), nil
}

// Verify compares a plaintext password with a hashed password
// Returns true if they match, false if they don't.
// Returns an error if there's a system error or if inputs are invalid
func Verify(plaintext, hashedPassword string) (bool, error) {
	if plaintext == "" {
		return false, ErrEmptyPassword
	}
	if hashedPassword == "" {
		return false, ErrEmptyHash
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plaintext))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, fmt.Errorf("password verify: %w", err)
	}

	return true, nil
}

// Cost returns the cost factor used to create the given hash
// Returns an error if the hash is invalid or empty
func Cost(hashedPassword string) (int, error) {
	if hashedPassword == "" {
		return 0, ErrEmptyHash
	}

	hash := []byte(hashedPassword)
	cost, err := bcrypt.Cost(hash)
	if err != nil {
		return 0, fmt.Errorf("password cost: %w", err)
	}

	return cost, nil
}
