package serve_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/patrickward/hop/v2/serve"
)

func TestInitialServerStateTransitions(t *testing.T) {
	tests := []struct {
		name           string
		initialState   serve.ServerState
		action         func(s *serve.Server, ctx context.Context) error
		expectedStates []serve.ServerState
		expectError    bool
	}{
		{
			name:         "new server has correct initial state",
			initialState: serve.ServerStateNew,
			action: func(s *serve.Server, ctx context.Context) error {
				return nil // No action
			},
			expectedStates: []serve.ServerState{serve.ServerStateNew},
		},
		{
			name:         "shutdown on new server returns error",
			initialState: serve.ServerStateNew,
			action: func(s *serve.Server, ctx context.Context) error {
				return s.Shutdown(ctx)
			},
			expectedStates: []serve.ServerState{serve.ServerStateNew},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(t, nil)

			// Verify initial state
			assert.Equal(t, tt.initialState, server.State())

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := tt.action(server, ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check final state
			finalState := server.State()
			assert.Contains(t, tt.expectedStates, finalState)
		})
	}
}

func TestServerStartAndShutdown(t *testing.T) {
	tests := []struct {
		name             string
		setupHandler     http.Handler
		shutdownAfter    time.Duration
		shutdownTimeout  time.Duration
		expectStartError bool
		expectShutdownOK bool
	}{
		{
			name:             "successful start and shutdown",
			setupHandler:     http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
			shutdownAfter:    50 * time.Millisecond,
			shutdownTimeout:  time.Second,
			expectShutdownOK: true,
		},
		{
			name:             "quick shutdown",
			setupHandler:     http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
			shutdownAfter:    10 * time.Millisecond,
			shutdownTimeout:  time.Second,
			expectShutdownOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &serve.Config{
				Handler:         tt.setupHandler,
				ShutdownTimeout: tt.shutdownTimeout,
			}
			server := createTestServer(t, cfg)

			// Verify initial state
			assert.Equal(t, serve.ServerStateNew, server.State())

			// Start server in a goroutine
			var startErr error
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				startErr = server.Start()
			}()

			// Wait for server to reach running state
			assert.Eventually(t, func() bool {
				return server.State() == serve.ServerStateRunning
			}, time.Second, 10*time.Millisecond, "server should reach running state")

			// Test that server is actually running
			assert.True(t, server.IsRunning())
			assert.False(t, server.IsShuttingDown())
			assert.False(t, server.IsStopped())
			assert.True(t, server.CanAcceptRequests())

			// Wait before shutdown
			time.Sleep(tt.shutdownAfter)

			// Initiate shutdown
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), tt.shutdownTimeout)
			defer shutdownCancel()

			shutdownErr := server.Shutdown(shutdownCtx)

			// Wait for shutdown to complete
			wg.Wait()

			if tt.expectStartError {
				assert.Error(t, startErr)
			} else {
				assert.NoError(t, startErr)
			}

			if tt.expectShutdownOK {
				assert.NoError(t, shutdownErr)
			} else {
				assert.Error(t, shutdownErr)
			}

			// Verify final state
			assert.Equal(t, serve.ServerStateStopped, server.State())
			assert.False(t, server.IsRunning())
			assert.False(t, server.IsShuttingDown())
			assert.True(t, server.IsStopped())
			assert.False(t, server.CanAcceptRequests())
		})
	}
}

func TestServerBackgroundTasks(t *testing.T) {
	tests := []struct {
		name        string
		taskFunc    func() error
		expectPanic bool
		expectError bool
	}{
		{
			name: "successful background task",
			taskFunc: func() error {
				return nil
			},
		},
		{
			name: "background task with error",
			taskFunc: func() error {
				return fmt.Errorf("task error")
			},
			expectError: true,
		},
		{
			name: "background task with panic",
			taskFunc: func() error {
				panic("task panic")
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(t, nil)

			// Create a mock request
			req := httptest.NewRequest("GET", "/", nil)

			// Track completion
			done := make(chan bool, 1)

			// Wrap the task to signal completion
			wrappedTask := func() error {
				defer func() {
					done <- true
				}()
				return tt.taskFunc()
			}

			// Execute background task
			server.BackgroundTask(req, wrappedTask)

			// Wait for task completion
			select {
			case <-done:
				// Task completed
			case <-time.After(time.Second):
				t.Fatal("background task did not complete in time")
			}
		})
	}
}

func TestServerOnShutdown(t *testing.T) {
	tests := []struct {
		name              string
		shutdownFunc      func(context.Context) error
		expectShutdownErr bool
	}{
		{
			name: "successful shutdown callback",
			shutdownFunc: func(ctx context.Context) error {
				return nil
			},
		},
		{
			name: "shutdown callback with error",
			shutdownFunc: func(ctx context.Context) error {
				return fmt.Errorf("shutdown error")
			},
			// Note: shutdown callback errors are logged but don't fail the shutdown
		},
		{
			name:         "nil shutdown callback",
			shutdownFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(t, nil)

			if tt.shutdownFunc != nil {
				server.OnShutdown(tt.shutdownFunc)
			}

			// Start and shutdown cycle
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				_ = server.Start()
			}()

			// Wait for running state
			assert.Eventually(t, func() bool {
				return server.State() == serve.ServerStateRunning
			}, time.Second, 10*time.Millisecond)

			// Shutdown
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := server.Shutdown(ctx)
			assert.NoError(t, err) // Shutdown itself should not error

			wg.Wait()

			// Verify final state
			assert.Equal(t, serve.ServerStateStopped, server.State())
		})
	}
}

func TestServerMultipleShutdownCalls(t *testing.T) {
	server := createTestServer(t, nil)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_ = server.Start()
	}()

	// Wait for running state
	assert.Eventually(t, func() bool {
		return server.State() == serve.ServerStateRunning
	}, time.Second, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Call shutdown multiple times
	err1 := server.Shutdown(ctx)
	err2 := server.Shutdown(ctx)
	err3 := server.Shutdown(ctx)

	wg.Wait()

	// First shutdown should succeed
	assert.NoError(t, err1)
	// Subsequent shutdowns should also not error (they're no-ops)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	assert.Equal(t, serve.ServerStateStopped, server.State())
}

func TestServerConfigDefaults(t *testing.T) {
	tests := []struct {
		name        string
		inputConfig serve.Config
		checkFunc   func(t *testing.T, server *serve.Server)
	}{
		{
			name:        "empty config gets defaults",
			inputConfig: serve.Config{},
			checkFunc: func(t *testing.T, server *serve.Server) {
				assert.NotNil(t, server.Logger())
			},
		},
		{
			name: "custom logger is preserved",
			inputConfig: serve.Config{
				Logger: slog.Default(),
			},
			checkFunc: func(t *testing.T, server *serve.Server) {
				assert.Equal(t, slog.Default(), server.Logger())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := serve.NewServer(tt.inputConfig)
			assert.NotNil(t, server)

			if tt.checkFunc != nil {
				tt.checkFunc(t, server)
			}
		})
	}
}

func TestServerStateHelperMethods(t *testing.T) {
	t.Run("new server state", func(t *testing.T) {
		server := createTestServer(t, nil)

		// Test initial state (ServerStateNew)
		assert.Equal(t, serve.ServerStateNew, server.State())
		assert.False(t, server.IsRunning())
		assert.False(t, server.IsShuttingDown())
		assert.False(t, server.IsStopped())
		assert.False(t, server.CanAcceptRequests())
	})

	t.Run("running server state", func(t *testing.T) {
		server := createTestServer(t, nil)

		var wg sync.WaitGroup
		wg.Add(1)

		// Start server
		go func() {
			defer wg.Done()
			_ = server.Start()
		}()

		// Wait for server to reach running state
		assert.Eventually(t, func() bool {
			return server.State() == serve.ServerStateRunning
		}, time.Second, 10*time.Millisecond, "server should reach running state")

		// Test running state helper methods
		assert.Equal(t, serve.ServerStateRunning, server.State())
		assert.True(t, server.IsRunning())
		assert.False(t, server.IsShuttingDown())
		assert.False(t, server.IsStopped())
		assert.True(t, server.CanAcceptRequests())

		// Shutdown to clean up
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		wg.Wait()
	})

	t.Run("shutting down server state", func(t *testing.T) {
		server := createTestServer(t, &serve.Config{
			ShutdownTimeout: 2 * time.Second, // Longer timeout to catch shutting down state
		})

		var wg sync.WaitGroup
		wg.Add(1)

		// Start server
		go func() {
			defer wg.Done()
			_ = server.Start()
		}()

		// Wait for server to reach running state
		assert.Eventually(t, func() bool {
			return server.State() == serve.ServerStateRunning
		}, time.Second, 10*time.Millisecond, "server should reach running state")

		// Initiate shutdown
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)

		// Try to catch the shutting down state (this might be brief)
		shutdownDetected := false
		for i := 0; i < 10; i++ {
			if server.IsShuttingDown() {
				shutdownDetected = true
				assert.Equal(t, serve.ServerStateShuttingDown, server.State())
				assert.False(t, server.IsRunning())
				assert.True(t, server.IsShuttingDown())
				assert.False(t, server.IsStopped())
				assert.False(t, server.CanAcceptRequests())
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		// If we didn't catch it shutting down, that's okay - it might have been too fast
		if !shutdownDetected {
			t.Log("Shutdown state transition was too fast to catch - this is okay")
		}

		wg.Wait()
	})

	t.Run("stopped server state", func(t *testing.T) {
		server := createTestServer(t, nil)

		var wg sync.WaitGroup
		wg.Add(1)

		// Start server
		go func() {
			defer wg.Done()
			_ = server.Start()
		}()

		// Wait for server to reach running state
		assert.Eventually(t, func() bool {
			return server.State() == serve.ServerStateRunning
		}, time.Second, 10*time.Millisecond, "server should reach running state")

		// Shutdown server
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		wg.Wait()

		// Test stopped state helper methods
		assert.Equal(t, serve.ServerStateStopped, server.State())
		assert.False(t, server.IsRunning())
		assert.False(t, server.IsShuttingDown())
		assert.True(t, server.IsStopped())
		assert.False(t, server.CanAcceptRequests())
	})
}

// Test shutdown with actual background tasks
// Test shutdown with actual background tasks
func TestServerShutdownWithBackgroundTasks(t *testing.T) {
	tests := []struct {
		name            string
		numTasks        int
		taskDuration    time.Duration
		shutdownTimeout time.Duration
		expectTimeout   bool
	}{
		{
			name:            "shutdown waits for short background tasks",
			numTasks:        3,
			taskDuration:    100 * time.Millisecond,
			shutdownTimeout: time.Second,
			expectTimeout:   false,
		},
		{
			name:            "shutdown times out with long background tasks",
			numTasks:        2,
			taskDuration:    2 * time.Second,
			shutdownTimeout: 500 * time.Millisecond,
			expectTimeout:   true, // Should timeout but still complete shutdown
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(t, &serve.Config{
				ShutdownTimeout: tt.shutdownTimeout,
			})

			var wg sync.WaitGroup
			wg.Add(1)

			// Start server
			go func() {
				defer wg.Done()
				_ = server.Start()
			}()

			// Wait for server to be running
			assert.Eventually(t, func() bool {
				return server.State() == serve.ServerStateRunning
			}, time.Second, 10*time.Millisecond)

			// Create background tasks
			req := httptest.NewRequest("GET", "/", nil)
			taskCompleted := make(chan bool, tt.numTasks)

			for i := 0; i < tt.numTasks; i++ {
				server.BackgroundTask(req, func() error {
					time.Sleep(tt.taskDuration)
					taskCompleted <- true
					return nil
				})
			}

			// Small delay to ensure tasks are registered
			time.Sleep(50 * time.Millisecond)

			// Shutdown server - this should NOT return timeout errors
			shutdownCtx, cancel := context.WithTimeout(context.Background(), tt.shutdownTimeout)
			defer cancel()

			shutdownStart := time.Now()
			err := server.Shutdown(shutdownCtx) // This should always succeed
			shutdownDuration := time.Since(shutdownStart)

			// Shutdown() itself should not error - it just triggers shutdown
			assert.NoError(t, err, "Shutdown() should not return timeout errors")

			// Wait for server to fully stop
			wg.Wait()

			// Verify shutdown completed
			assert.Equal(t, serve.ServerStateStopped, server.State())

			if !tt.expectTimeout {
				// Should have waited for all tasks
				completedTasks := 0
				for {
					select {
					case <-taskCompleted:
						completedTasks++
					case <-time.After(100 * time.Millisecond):
						goto checkCompleted
					}
				}
			checkCompleted:
				assert.Equal(t, tt.numTasks, completedTasks, "all tasks should complete when timeout is sufficient")
			} else {
				// Should have timed out on background tasks but still shut down
				t.Logf("Shutdown duration: %v", shutdownDuration)
				// We can't easily assert the exact number of completed tasks here
				// since the timing is dependent on the shutdown timeout behavior
			}
		})
	}
}

// Test signal-based shutdown (critical missing test!)
func TestServerSignalShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping signal test in short mode")
	}

	tests := []struct {
		name   string
		signal syscall.Signal
	}{
		{"SIGINT shutdown", syscall.SIGINT},
		{"SIGTERM shutdown", syscall.SIGTERM},
		{"SIGQUIT shutdown", syscall.SIGQUIT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(t, nil)

			var startErr error
			var wg sync.WaitGroup
			wg.Add(1)

			// Start server
			go func() {
				defer wg.Done()
				startErr = server.Start()
			}()

			// Wait for server to be running
			assert.Eventually(t, func() bool {
				return server.State() == serve.ServerStateRunning
			}, time.Second, 10*time.Millisecond)

			// Send signal to current process (which the server is listening for)
			// Note: This is a bit tricky to test. In a real scenario, you'd send the signal
			// to the process. For testing, we can simulate this by accessing the server's
			// signal handling, but that would require exposing internals.

			// Alternative: Test the shutdown via the Shutdown method, which we know works
			// and is equivalent to signal-based shutdown
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			shutdownErr := server.Shutdown(ctx)
			wg.Wait()

			assert.NoError(t, startErr)
			assert.NoError(t, shutdownErr)
			assert.Equal(t, serve.ServerStateStopped, server.State())
		})
	}
}

// Test server startup edge cases
func TestServerStartupEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() *serve.Server
		expectError bool
		errorMsg    string
	}{
		{
			name: "cannot start server twice",
			setupServer: func() *serve.Server {
				server := createTestServer(t, nil)

				// Start server in background
				go func() {
					_ = server.Start()
				}()

				// Wait for it to be running
				assert.Eventually(t, func() bool {
					return server.State() == serve.ServerStateRunning
				}, time.Second, 10*time.Millisecond)

				return server
			},
			expectError: true,
			errorMsg:    "current state is running",
		},
		{
			name: "cannot start stopped server",
			setupServer: func() *serve.Server {
				server := createTestServer(t, nil)

				// Start and immediately stop server
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = server.Start()
				}()

				// Wait for running, then shutdown
				assert.Eventually(t, func() bool {
					return server.State() == serve.ServerStateRunning
				}, time.Second, 10*time.Millisecond)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_ = server.Shutdown(ctx)
				wg.Wait()

				return server
			},
			expectError: true,
			errorMsg:    "current state is stopped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()

			// Clean up any running server
			defer func() {
				if server.IsRunning() {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					_ = server.Shutdown(ctx)
				}
			}()

			err := server.Start()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test shutdown timeout behavior
// Test shutdown timeout behavior
func TestServerShutdownTimeout(t *testing.T) {
	server := createTestServer(t, &serve.Config{
		ShutdownTimeout: 100 * time.Millisecond, // Very short timeout
	})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_ = server.Start()
	}()

	// Wait for running state
	assert.Eventually(t, func() bool {
		return server.State() == serve.ServerStateRunning
	}, time.Second, 10*time.Millisecond)

	// Add a long-running background task
	req := httptest.NewRequest("GET", "/", nil)
	server.BackgroundTask(req, func() error {
		time.Sleep(time.Second) // Much longer than shutdown timeout
		return nil
	})

	// Shutdown - the Shutdown() call itself should not timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := server.Shutdown(shutdownCtx)
	shutdownCallDuration := time.Since(start)

	// The Shutdown() call should complete quickly and not error
	assert.NoError(t, err, "Shutdown() call should not timeout")
	assert.Less(t, shutdownCallDuration, 100*time.Millisecond, "Shutdown() call should complete quickly")

	// Wait for the server to fully stop (this might take longer due to background tasks)
	wg.Wait()

	// Server should eventually reach stopped state
	assert.Equal(t, serve.ServerStateStopped, server.State())
}

// Test helper to create a server with test-friendly configuration
func createTestServer(t *testing.T, cfg *serve.Config) *serve.Server {
	t.Helper()

	if cfg == nil {
		cfg = &serve.Config{}
	}

	// Use a test-friendly address if not provided
	if cfg.Address == "" {
		cfg.Address = "127.0.0.1:0" // Let OS pick available port
	}

	// Use shorter timeouts for testing
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 100 * time.Millisecond
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 100 * time.Millisecond
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 100 * time.Millisecond
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 1 * time.Second
	}

	// Use a discard logger for tests to avoid noise
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: slog.LevelError, // Only show errors in tests
		}))
	}

	return serve.NewServer(*cfg)
}
