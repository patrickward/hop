package token_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/secure/token"
)

func TestTokenStore(t *testing.T) {
	t.Run("create and get token", func(t *testing.T) {
		store := token.NewTokenStore(nil) // Use default options
		defer store.Close()

		message := "test message"
		newToken, err := store.Create(message)
		require.NoError(t, err)
		require.NotEmpty(t, newToken)

		// Verify token exists
		valid, err := store.Verify(newToken)
		require.NoError(t, err)
		assert.True(t, valid)

		// Get message
		got, err := store.Get(newToken)
		require.NoError(t, err)
		assert.Equal(t, message, got)

		// Token should be consumed
		_, err = store.Get(newToken)
		assert.ErrorIs(t, err, token.ErrTokenNotFound)
	})

	t.Run("token expiration", func(t *testing.T) {
		// Create store with short expiration but LONG cleanup interval
		opts := &token.Options{
			Expiration:      1 * time.Second,
			CleanupInterval: 1 * time.Hour, // Long cleanup to avoid interference
		}
		store := token.NewTokenStore(opts)
		defer store.Close()

		message := "expiring message"
		newToken, err := store.Create(message)
		require.NoError(t, err)

		// Verify token is valid initially
		valid, err := store.Verify(newToken)
		require.NoError(t, err)
		assert.True(t, valid)

		// Wait for token to expire
		time.Sleep(2 * time.Second)

		// Try to get expired token
		_, err = store.Get(newToken)
		assert.ErrorIs(t, err, token.ErrTokenExpired)

		// Verify should also return false for expired token
		valid, err = store.Verify(newToken)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("invalid token", func(t *testing.T) {
		store := token.NewTokenStore(nil)
		defer store.Close()

		_, err := store.Get("invalid_token")
		assert.ErrorIs(t, err, token.ErrTokenNotFound)
	})

	t.Run("verify token", func(t *testing.T) {
		store := token.NewTokenStore(nil)
		defer store.Close()

		message := "verify test"
		newToken, err := store.Create(message)
		require.NoError(t, err)

		// Verify valid token
		valid, err := store.Verify(newToken)
		require.NoError(t, err)
		assert.True(t, valid)

		// Verify invalid token
		valid, err = store.Verify("invalid_token")
		require.NoError(t, err)
		assert.False(t, valid)

		// Get token should still work after verify
		got, err := store.Get(newToken)
		require.NoError(t, err)
		assert.Equal(t, message, got)
	})

	t.Run("cleanup removes expired tokens", func(t *testing.T) {
		// Create store with short expiration and cleanup interval
		opts := &token.Options{
			Expiration:      1 * time.Second,
			CleanupInterval: 2 * time.Second,
		}
		store := token.NewTokenStore(opts)
		defer store.Close()

		message := "cleanup test"
		newToken, err := store.Create(message)
		require.NoError(t, err)

		// Wait for cleanup to run
		time.Sleep(3 * time.Second)

		// Token should be removed by cleanup
		_, err = store.Get(newToken)
		assert.ErrorIs(t, err, token.ErrTokenNotFound)
	})
}
