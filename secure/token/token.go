package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// EncoderBase32 specifies base32 encoding for token generation
	EncoderBase32 = "base32"
	// EncoderHex specifies hexadecimal encoding for token generation
	EncoderHex = "hex"

	// DefaultLength is the default token length in bytes
	DefaultLength = 16
	// DefaultEncoder is the default encoding method
	DefaultEncoder = EncoderBase32

	// MinLength is the minimum allowed token length in bytes
	MinLength = 16
	// MaxLength is the maximum allowed token length in bytes
	MaxLength = 64
)

type config struct {
	length  int
	encoder string
}

type Option func(*config)

// WithLength sets the token length in bytes
// If length is outside valid range (MinLength to MaxLength),
// DefaultLength will be used instead
func WithLength(length int) Option {
	return func(c *config) {
		if length >= MinLength && length <= MaxLength {
			c.length = length
		}
	}
}

// WithEncoder sets the encoding format (EncoderBase32 or EncoderHex)
// If an invalid encoder is specified, DefaultEncoder will be used
func WithEncoder(encoder string) Option {
	return func(c *config) {
		switch encoder {
		case EncoderBase32, EncoderHex:
			c.encoder = encoder
		}
	}
}

func defaultConfig() *config {
	return &config{
		length:  DefaultLength,
		encoder: DefaultEncoder,
	}
}

// Generate creates a new random token with the given options
// If no options are provided, DefaultLength and DefaultEncoder will be used
func Generate(opts ...Option) (string, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(cfg)
	}

	b := make([]byte, cfg.length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("token generate: %w", err)
	}

	// If encoder is invalid, fall back to default
	if cfg.encoder != EncoderBase32 && cfg.encoder != EncoderHex {
		cfg.encoder = DefaultEncoder
	}

	switch cfg.encoder {
	case EncoderBase32:
		return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)), nil
	case EncoderHex:
		return hex.EncodeToString(b), nil
	default:
		// This shouldn't happen due to the fallback above, but included for completeness
		return "", fmt.Errorf("unsupported encoding: %s", cfg.encoder)
	}
}

// Hash creates a SHA-256 hash of the token
// Returns the hash as a lowercase hex-encoded string
func Hash(plaintext string) string {
	hash := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(hash[:])
}
