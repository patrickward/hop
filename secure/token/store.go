package token

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrTokenGeneration = errors.New("failed to generate token")
	ErrTokenNotFound   = errors.New("token not found")
	ErrTokenExpired    = errors.New("token expired")
)

// Store is a thread-safe store for tokens
type Store struct {
	sync.RWMutex                      // Protects the tokens map
	done         chan struct{}        // Used to signal the cleanup goroutine to stop
	tokens       map[string]tokenData // Map of hashed tokens to tokenData
	closeOnce    sync.Once            // Ensures the store is closed only once
	expiration   time.Duration        // How long tokens are valid for
	cleanupInt   time.Duration        // How often to run cleanup
}

// tokenData holds the message and creation time for a token
type tokenData struct {
	message   string
	createdAt time.Time
}

// Options configures the Store
type Options struct {
	Expiration      time.Duration // Token expiration duration
	CleanupInterval time.Duration // Cleanup interval duration
}

// DefaultOptions provides recommended settings
var DefaultOptions = Options{
	Expiration:      10 * time.Minute,
	CleanupInterval: 5 * time.Minute,
}

// NewTokenStore creates a new token store with the given options
func NewTokenStore(opts *Options) *Store {
	if opts == nil {
		opts = &DefaultOptions
	}

	ts := &Store{
		tokens:     make(map[string]tokenData),
		done:       make(chan struct{}),
		expiration: opts.Expiration,
		cleanupInt: opts.CleanupInterval,
	}

	// Start a goroutine to clean up expired tokens periodically
	go ts.cleanup()

	return ts
}

// Close closes the store and cleans up any resources
func (ts *Store) Close() {
	ts.closeOnce.Do(func() {
		close(ts.done)
	})
}

// Create creates a new token with the given message and returns the token
func (ts *Store) Create(message string) (string, error) {
	// Generate a new token using our secure token package
	plainToken, err := Generate()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	// Hash the token for storage
	hashedToken := Hash(plainToken)

	ts.Lock()
	ts.tokens[hashedToken] = tokenData{
		message:   message,
		createdAt: time.Now(),
	}
	ts.Unlock()

	return plainToken, nil
}

// Get returns the message associated with the given token
// The token is deleted after retrieval (one-time use)
func (ts *Store) Get(plainToken string) (string, error) {
	hashedToken := Hash(plainToken)

	ts.Lock()
	defer ts.Unlock()

	data, found := ts.tokens[hashedToken]
	if !found {
		return "", ErrTokenNotFound
	}

	// Check if token has expired
	if time.Since(data.createdAt) > ts.expiration {
		delete(ts.tokens, hashedToken) // Clean up expired token
		return "", ErrTokenExpired
	}

	// Delete the token after use
	delete(ts.tokens, hashedToken)
	return data.message, nil
}

// Verify checks if a token exists and is valid without consuming it
func (ts *Store) Verify(plainToken string) (bool, error) {
	hashedToken := Hash(plainToken)

	ts.RLock()
	defer ts.RUnlock()

	data, found := ts.tokens[hashedToken]
	if !found {
		return false, nil
	}

	// Check if token has expired
	if time.Since(data.createdAt) > ts.expiration {
		return false, nil
	}

	return true, nil
}

func (ts *Store) cleanup() {
	ticker := time.NewTicker(ts.cleanupInt)
	defer ticker.Stop()

	for {
		select {
		case <-ts.done:
			return
		case <-ticker.C:
			ts.Lock()
			now := time.Now()
			for hashedToken, data := range ts.tokens {
				if now.Sub(data.createdAt) > ts.expiration {
					delete(ts.tokens, hashedToken)
				}
			}
			ts.Unlock()
		}
	}
}
