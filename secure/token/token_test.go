package token_test

import (
	"encoding/base32"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/secure/token"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		opts        []token.Option
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result string)
	}{
		{
			name: "default options",
			opts: nil,
			validate: func(t *testing.T, result string) {
				// Convert to uppercase for base32 decoding
				upperResult := strings.ToUpper(result)
				decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(upperResult)
				require.NoError(t, err)
				assert.Len(t, decoded, token.DefaultLength)
			},
		},
		{
			name: "custom length with base32",
			opts: []token.Option{
				token.WithLength(32),
				token.WithEncoder(token.EncoderBase32),
			},
			validate: func(t *testing.T, result string) {
				upperResult := strings.ToUpper(result)
				decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(upperResult)
				require.NoError(t, err)
				assert.Len(t, decoded, 32)
			},
		},
		{
			name: "hex encoder",
			opts: []token.Option{
				token.WithEncoder(token.EncoderHex),
			},
			validate: func(t *testing.T, result string) {
				decoded, err := hex.DecodeString(result)
				require.NoError(t, err)
				assert.Len(t, decoded, token.DefaultLength)
			},
		},
		{
			name: "invalid encoder",
			opts: []token.Option{
				token.WithEncoder("invalid"),
			},
			validate: func(t *testing.T, result string) {
				// Should fall back to default encoder (base32)
				upperResult := strings.ToUpper(result)
				decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(upperResult)
				require.NoError(t, err)
				assert.Len(t, decoded, token.DefaultLength)
			},
		},
		{
			name: "invalid length",
			opts: []token.Option{
				token.WithLength(-1),
			},
			validate: func(t *testing.T, result string) {
				// Should use default length
				upperResult := strings.ToUpper(result)
				decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(upperResult)
				require.NoError(t, err)
				assert.Len(t, decoded, token.DefaultLength)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := token.Generate(tt.opts...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		wantLen  int
		wantHash string
	}{
		{
			name:    "normal token",
			token:   "abcdef123456",
			wantLen: 64, // SHA-256 hex encoded length
		},
		{
			name:     "empty token",
			token:    "",
			wantLen:  64,
			wantHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "known token",
			token:    "test",
			wantLen:  64,
			wantHash: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := token.Hash(tt.token)

			assert.Len(t, result, tt.wantLen)

			// Verify it's valid hex
			_, err := hex.DecodeString(result)
			assert.NoError(t, err)

			if tt.wantHash != "" {
				assert.Equal(t, tt.wantHash, result)
			}

			// Verify deterministic behavior
			assert.Equal(t, result, token.Hash(tt.token))
		})
	}
}

func TestTokenUniqueness(t *testing.T) {
	// Generate multiple tokens and ensure they're unique
	tokens := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		token, err := token.Generate()
		require.NoError(t, err)

		// Ensure token is unique
		assert.False(t, tokens[token], "Token collision detected")
		tokens[token] = true
	}
}
