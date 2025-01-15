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

	"github.com/patrickward/hop/route"
)

// DataFunc is a function type that takes an HTTP request and a pointer to a map of data.
// It represents a callback function that can be used to populate data for templates.
type DataFunc func(r *http.Request, data *map[string]any)

type Server struct {
	address         string
	idleTimeout     time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
	shutdownTimeout time.Duration
	onShutdown      func(context.Context) error
	httpServer      *http.Server
	logger          *slog.Logger
	router          *route.Mux
	wg              *sync.WaitGroup
	stopChan        chan struct{}
	stopping        sync.Once
}

type Config struct {
	Address         string
	IdleTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	Router          *route.Mux
	Logger          *slog.Logger
}

// NewServer creates a new server with the given configuration and logger.
func NewServer(cfg Config) *Server {
	if cfg.Router == nil {
		cfg.Router = route.New()
	}

	httpServer := &http.Server{
		//Addr:         fmt.Sprintf(":%d", config.Server.Port),
		Addr:         cfg.Address,
		Handler:      cfg.Router,
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
		router:          cfg.Router,
		wg:              &sync.WaitGroup{},
		stopChan:        make(chan struct{}),
	}

	return srv
}

// Logger returns the logger for the server.
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// Router returns the router for the server.
func (s *Server) Router() *route.Mux {
	return s.router
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

		if err := s.httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	// Graceful shutdown handler
	eg.Go(func() error {
		<-gCtx.Done()

		s.logger.Info("initiating graceful shutdown")

		// Split the shutdown timeout between WaitGroup and server shutdown
		totalTimeout := s.shutdownTimeout
		wgTimeout := totalTimeout / 2
		serverTimeout := totalTimeout - wgTimeout

		// Wait for background tasks
		wgDone := make(chan struct{})
		go func() {
			s.logger.Info("waiting for background tasks to complete",
				slog.Duration("timeout", wgTimeout))
			s.wg.Wait()
			close(wgDone)
		}()

		// Create context for WaitGroup timeout
		wgCtx, wgCancel := context.WithTimeout(context.Background(), wgTimeout)
		defer wgCancel()

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

		// Create context for server shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			serverTimeout,
		)
		defer shutdownCancel()

		s.logger.Info("shutting down http server",
			slog.Duration("timeout", serverTimeout))

		// Proceed with server shutdown
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}

		return nil
	})

	// Wait for all errgroup goroutines to complete or error
	if err := eg.Wait(); err != nil &&
		!errors.Is(err, context.Canceled) {
		return fmt.Errorf("server error: %w", err)
	}

	s.logger.Info("server exited")
	return nil
}

// Shutdown initiates a graceful shutdown of the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Use sync.Once to ensure we only trigger shutdown once
	s.stopping.Do(func() {
		close(s.stopChan)
	})

	// Wait for the context to be done or a reasonable timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
		// Give a short grace period for shutdown to begin
		return nil
	}
}
