// Package serve provides a way to create and manage an HTTP server, including routing, middleware, and graceful shutdown.
package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type ServerState int

const (
	ServerStateNew          ServerState = iota + 1 // Server is newly created and not started
	ServerStateStarting                            // Server is starting up
	ServerStateRunning                             // Server is currently running
	ServerStateShuttingDown                        // Server is in the process of shutting down
	ServerStateStopped                             // Server has been stopped
)

func (s ServerState) String() string {
	switch s {
	case ServerStateNew:
		return "new"
	case ServerStateStarting:
		return "starting"
	case ServerStateRunning:
		return "running"
	case ServerStateShuttingDown:
		return "shutting_down"
	case ServerStateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// Server represents an HTTP server with configurable settings and graceful shutdown capabilities.
type Server struct {
	address         string
	idleTimeout     time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	shutdownTimeout time.Duration
	onShutdown      func(context.Context) error
	httpServer      *http.Server
	logger          *slog.Logger
	handler         http.Handler
	wg              *sync.WaitGroup
	stopChan        chan struct{}
	stopping        sync.Once
	state           ServerState
	stateMu         sync.RWMutex // Protects the server state
}

// Config holds the configuration for the server.
type Config struct {
	Address         string
	IdleTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Handler         http.Handler
	Logger          *slog.Logger
}

// NewServer creates a new server with the given configuration and logger.
func NewServer(cfg Config) *Server {
	if cfg.Handler == nil {
		cfg.Handler = http.DefaultServeMux
	}

	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 120 * time.Second
	}

	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 5 * time.Second
	}

	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}

	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}

	httpServer := &http.Server{
		//Addr:         fmt.Sprintf(":%d", config.Server.Port),
		Addr:         cfg.Address,
		Handler:      cfg.Handler,
		ErrorLog:     slog.NewLogLogger(cfg.Logger.Handler(), slog.LevelWarn),
		IdleTimeout:  cfg.IdleTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	srv := &Server{
		address:         cfg.Address,
		idleTimeout:     cfg.IdleTimeout,
		readTimeout:     cfg.ReadTimeout,
		writeTimeout:    cfg.WriteTimeout,
		shutdownTimeout: cfg.ShutdownTimeout,
		httpServer:      httpServer,
		logger:          cfg.Logger,
		handler:         cfg.Handler,
		wg:              &sync.WaitGroup{},
		stopChan:        make(chan struct{}),
		state:           ServerStateNew,
	}

	return srv
}

func (s *Server) State() ServerState {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.state
}

func (s *Server) setState(state ServerState) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()

	oldState := s.state
	s.state = state
	if oldState != state {
		s.logger.Info("server state changed",
			slog.String("old_state", oldState.String()),
			slog.String("new_state", state.String()))
	}
}

// Logger returns the logger for the server.
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// OnShutdown registers a shutdown handler to be called before the server stops
func (s *Server) OnShutdown(fn func(context.Context) error) {
	s.onShutdown = fn
}

// BackgroundTask runs a function in a goroutine, and reports any errors to the server's error logger.
func (s *Server) BackgroundTask(r *http.Request, fn func() error) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		defer func() {
			err := recover()
			if err != nil {
				s.ReportServerError(r, fmt.Errorf("%s", err))
			}
		}()

		err := fn()
		if err != nil {
			s.ReportServerError(r, err)
		}
	}()
}

// Start starts the server and listens for incoming requests. It will block until the server is shut down.
func (s *Server) Start() error {
	// Check if the server is already running or has been used
	if s.State() != ServerStateNew {
		return fmt.Errorf("server cannot be started: current state is %s", s.State())
	}

	s.setState(ServerStateStarting)

	// Create base context for signals
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Create context that can be canceled either by signals or stopChan
	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	// Handle both signal context and stopChan
	go func() {
		select {
		case <-ctx.Done():
			s.logger.Info("received shutdown signal",
				slog.String("cause", ctx.Err().Error()))
			runCancel()
		case <-s.stopChan:
			s.logger.Info("received shutdown request")
			runCancel()
		}
	}()

	// Create errgroup with our cancellable context
	eg, gCtx := errgroup.WithContext(runCtx)

	// Start HTTP server
	eg.Go(func() error {
		s.logger.Info("starting server",
			slog.Group("server", slog.String("addr", s.httpServer.Addr)))

		s.setState(ServerStateRunning)

		if err := s.httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			s.setState(ServerStateStopped)
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	// Graceful shutdown handler
	eg.Go(func() error {
		<-gCtx.Done()

		s.setState(ServerStateShuttingDown)
		s.logger.Info("initiating graceful shutdown")

		// Use a single context for the shutdown process
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer shutdownCancel()

		// Wait for background tasks with a portion of the shutdown timeout
		wgTimeout := s.shutdownTimeout / 2
		wgCtx, wgCancel := context.WithTimeout(shutdownCtx, wgTimeout)
		defer wgCancel()

		wgDone := make(chan struct{})
		go func() {
			s.logger.Info("waiting for background tasks to complete",
				slog.Duration("timeout", wgTimeout))
			s.wg.Wait()
			close(wgDone)
		}()

		// Wait for either WaitGroup completion or timeout
		select {
		case <-wgDone:
			s.logger.Info("all background tasks completed")
		case <-wgCtx.Done():
			s.logger.Warn("timeout waiting for background tasks",
				slog.Duration("elapsed", wgTimeout))
		}

		// Call onShutdown handler if registered
		if s.onShutdown != nil {
			if err := s.onShutdown(context.Background()); err != nil {
				s.logger.Error("onShutdown error", slog.String("error", err.Error()))
			}
		}

		// Use the remaining time for server shutdown
		s.logger.Info("shutting down http server")
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}

		s.setState(ServerStateStopped)
		return nil
	})

	// Wait for all goroutines to complete or error
	err := eg.Wait()

	// Ensure the server state is set to `stopped` if it wasn't already
	if s.State() != ServerStateStopped {
		s.setState(ServerStateStopped)
	}

	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server error: %w", err)
	}

	s.logger.Info("server exited")
	return nil
}

// Shutdown initiates a graceful shutdown of the server
func (s *Server) Shutdown(ctx context.Context) error {
	currentState := s.State()

	if currentState >= ServerStateShuttingDown {
		s.logger.Warn("shutdown requested while already shutting down or stopped", slog.String("state", currentState.String()))
		return nil
	}

	if currentState != ServerStateRunning {
		return fmt.Errorf("server cannot be shut down: current state is %s", currentState)
	}

	s.logger.Info("shutdown requested", slog.String("state", currentState.String()))

	// Use sync.Once to ensure we only trigger shutdown once
	s.stopping.Do(func() {
		close(s.stopChan)
	})

	// Don't wait for the context or return its error - just trigger the shutdown process
	// The actual shutdown will be handled in the Start method's goroutine
	return nil
}

// IsRunning returns true if the server is currently running
func (s *Server) IsRunning() bool {
	return s.State() == ServerStateRunning
}

// IsShuttingDown returns true if the server is in the process of shutting down
func (s *Server) IsShuttingDown() bool {
	return s.State() == ServerStateShuttingDown
}

// IsStopped returns true if the server has stopped
func (s *Server) IsStopped() bool {
	return s.State() == ServerStateStopped
}

// CanAcceptRequests returns true if the server can accept new requests
func (s *Server) CanAcceptRequests() bool {
	state := s.State()
	return state == ServerStateRunning
}
