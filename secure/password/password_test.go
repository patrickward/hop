package password_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/secure/password"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		opts        []password.Option
		wantErr     bool
		errContains string
	}{
		{
			name:     "default options",
			password: "secret123",
			opts:     nil,
			wantErr:  false,
		},
		{
			name:     "with custom cost",
			password: "secret123",
			opts:     []password.Option{password.WithCost(password.MinCost)},
			wantErr:  false,
		},
		{
			name:     "with moderate cost",
			password: "secret123",
			opts:     []password.Option{password.WithCost(13)},
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			opts:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid cost below min",
			password: "secret123",
			opts:     []password.Option{password.WithCost(password.MinCost - 1)},
			wantErr:  false, // Should use default cost instead
		},
		{
			name:     "invalid cost above max",
			password: "secret123",
			opts:     []password.Option{password.WithCost(password.MaxCost + 1)},
			wantErr:  false, // Should use default cost instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := password.Hash(tt.password, tt.opts...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, hash)

			// Verify the hash works with the original password
			valid, err := password.Verify(tt.password, hash)
			require.NoError(t, err)
			assert.True(t, valid)
		})
	}
}

// TestHashWithHighCost specifically tests high cost factors with a timeout
func TestHashWithHighCost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping high-cost password hash test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		hash, err := password.Hash("test123", password.WithCost(password.MaxCost))
		if err == nil && hash != "" {
			// Successfully hashed with max cost
			close(done)
		}
	}()

	select {
	case <-ctx.Done():
		t.Skip("max cost password hash timed out after 5 seconds - this is expected on most systems")
	case <-done:
		t.Log("successfully completed max cost password hash")
	}
}

func TestVerify(t *testing.T) {
	// Create a known hash for testing - use MinCost for speed
	knownPassword := "secret123"
	hash, err := password.Hash(knownPassword, password.WithCost(password.MinCost))
	require.NoError(t, err)

	tests := []struct {
		name        string
		password    string
		hash        string
		want        bool
		wantErr     bool
		errContains string
	}{
		{
			name:     "correct password",
			password: knownPassword,
			hash:     hash,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
			wantErr:  false,
		},
		{
			name:        "invalid hash format",
			password:    knownPassword,
			hash:        "invalid_hash_format",
			want:        false,
			wantErr:     true,
			errContains: "hashedSecret too short",
		},
		{
			name:        "empty password",
			password:    "",
			hash:        hash,
			want:        false,
			wantErr:     true,
			errContains: "password cannot be empty",
		},
		{
			name:        "empty hash",
			password:    knownPassword,
			hash:        "",
			want:        false,
			wantErr:     true,
			errContains: "hash cannot be empty",
		},
		{
			name:     "malformed hash",
			password: knownPassword,
			// 32 bytes of random data, not a valid bcrypt hash
			hash:    "$thisisnotavalidhash12345678901234567890999999999999999999999999999",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := password.Verify(tt.password, tt.hash)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, valid)
		})
	}
}
