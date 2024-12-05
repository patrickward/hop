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

	"golang.org/x/sync/errgroup"

	"github.com/patrickward/hop/conf"
	"github.com/patrickward/hop/route"
)

// DataFunc is a function type that takes an HTTP request and a pointer to a map of data.
// It represents a callback function that can be used to populate data for templates.
type DataFunc func(r *http.Request, data *map[string]any)

type Server struct {
	config     *conf.Config
	onShutdown func(context.Context) error
	httpServer *http.Server
	logger     *slog.Logger
	router     *route.Mux
	wg         *sync.WaitGroup
}

// NewServer creates a new server with the given configuration and logger.
func NewServer(config *conf.Config, logger *slog.Logger, router *route.Mux) *Server {
	if router == nil {
		router = route.New()
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Server.Port),
		Handler:      router,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
		IdleTimeout:  config.Server.IdleTimeout.Duration,
		ReadTimeout:  config.Server.ReadTimeout.Duration,
		WriteTimeout: config.Server.WriteTimeout.Duration,
	}

	srv := &Server{
		config:     config,
		httpServer: httpServer,
		logger:     logger,
		router:     router,
		wg:         &sync.WaitGroup{},
	}

	return srv
}

// Config returns the server configuration.
func (s *Server) Config() *conf.Config {
	return s.config
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
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Crete errgroup with signal context
	eg, gCtx := errgroup.WithContext(ctx)

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

		s.logger.Info("initiating graceful shutdown",
			slog.String("cause", gCtx.Err().Error()))

		// Split the shutdown timeout between WaitGroup and server shutdown
		totalTimeout := s.config.Server.ShutdownTimeout.Duration
		wgTimeout := totalTimeout / 2
		serverTimeout := totalTimeout - wgTimeout

		// Create a channel to signal WaitGroup completion
		wgDone := make(chan struct{})

		// Wait for background tasks in a separate goroutine
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

		// Create context for server shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			serverTimeout,
		)
		defer shutdownCancel()

		// Call onShutdown handler if registered
		if s.onShutdown != nil {
			if err := s.onShutdown(shutdownCtx); err != nil {
				s.logger.Error("onShutdown error", slog.String("error", err.Error()))
			}
		}

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
