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

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
	"golang.org/x/sync/errgroup"

	"github.com/patrickward/hop/conf"
	"github.com/patrickward/hop/render"
	"github.com/patrickward/hop/wrap"
)

// CleanupFunc is a function type that can be used for cleanup tasks when the server shuts down.
type CleanupFunc func()

// DataFunc is a function type that takes an HTTP request and a pointer to a map of data.
// It represents a callback function that can be used to populate data for templates.
type DataFunc func(r *http.Request, data *map[string]any)

type Server struct {
	config      *conf.BaseConfig
	cleanupFunc CleanupFunc
	dataFunc    DataFunc
	httpServer  *http.Server
	logger      *slog.Logger
	router      *http.ServeMux
	tm          *render.TemplateManager
	session     *scs.SessionManager
	wg          *sync.WaitGroup
}

// NewServer creates a new server with the given configuration and logger.
func NewServer(config *conf.BaseConfig, logger *slog.Logger, tm *render.TemplateManager, session *scs.SessionManager) *Server {
	router := http.NewServeMux()
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
		session:    session,
		tm:         tm,
		wg:         &sync.WaitGroup{},
	}

	return srv
}

// BaseConfig returns the server configuration.
func (s *Server) BaseConfig() *conf.BaseConfig {
	return s.config
}

// Logger returns the logger for the server.
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// Session returns the session manager for the server.
func (s *Server) Session() *scs.SessionManager {
	return s.session
}

// TM returns the template manager for the server.
func (s *Server) TM() *render.TemplateManager {
	return s.tm
}

// AddRoute adds a new route to the server, using the newer v1.22 http.Handler interface. It takes a pattern, an http.Handler, and an optional list of middleware.
func (s *Server) AddRoute(pattern string, handler http.Handler, middleware ...wrap.Middleware) {
	if len(middleware) > 0 {
		// Create a chain of middleware and wrap the handler
		chain := wrap.New(middleware...).Then(handler)
		s.router.Handle(pattern, chain)
		return
	}
	s.router.Handle(pattern, handler)
}

// AddChainedRoute adds a new route to the server with a chain of middleware
// It takes a pattern, an http.Handler, and a wrap.Chain struct
func (s *Server) AddChainedRoute(pattern string, handler http.Handler, chain wrap.Chain) {
	s.router.Handle(pattern, chain.Then(handler))
}

// AddRoutes adds multiple routes to the server. It takes a map of patterns to http.Handlers and an optional list of middleware.
func (s *Server) AddRoutes(routes map[string]http.Handler, middleware ...wrap.Middleware) {
	for pattern, handler := range routes {
		if len(middleware) > 0 {
			s.AddRoute(pattern, handler, middleware...)
			continue
		}
		s.AddRoute(pattern, handler)
	}
}

// AddChainedRoutes adds multiple routes to the server with a chain of middleware
func (s *Server) AddChainedRoutes(routes map[string]http.Handler, chain wrap.Chain) {
	for pattern, handler := range routes {
		s.AddChainedRoute(pattern, handler, chain)
	}
}

// RegisterCleanup registers a cleanup function to be called when the server shuts down.
func (s *Server) RegisterCleanup(fn func()) {
	s.cleanupFunc = fn
}

// RegisterTemplateData registers a function that populates template data each time a template is rendered.
func (s *Server) RegisterTemplateData(fn DataFunc) {
	s.dataFunc = fn
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

// CacheBuster returns a string that can be used to bust the cache on static assets.
func (s *Server) CacheBuster() string {
	return time.Now().Format("20060102150405")
}

// NewTemplateData returns a map of data that can be used in a Go template. It includes the current user, environment, version, and other useful information.
func (s *Server) NewTemplateData(r *http.Request) map[string]any {
	// Check if this is the home page.
	isHome := r.URL.Path == "/"

	//prodCSS := strings.TrimSpace(string(assets.CSSHash))
	//prodJS := strings.TrimSpace(string(assets.JSHash))
	//prodCSSFile := fmt.Sprintf("/static/css/app.min.%s.css", prodCSS)
	//prodJSFile := fmt.Sprintf("/static/js/app.min.%s.js", prodJS)

	data := map[string]any{
		//"CurrentUser":        auth.GetCurrentUserFromContext(r),
		"Env":           s.config.Environment,
		"IsDevelopment": s.config.Environment == "development",
		"IsProduction":  s.config.Environment == "production",
		"CSRFToken":     nosurf.Token(r),
		"BaseURL":       s.config.Server.BaseURL,
		"CacheBuster":   s.CacheBuster,
		"RequestPath":   r.URL.Path,
		"IsHome":        isHome,
		//"ProdCSSFile":        prodCSSFile,
		//"ProdJSFile":         prodJSFile,
		"MaintenanceEnabled": s.config.Maintenance.Enabled,
		"MaintenanceMessage": s.config.Maintenance.Message,
	}

	if s.dataFunc != nil {
		s.dataFunc(r, &data)
	}

	return data
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

		s.logger.Info("shutting down http server",
			slog.Duration("timeout", serverTimeout))

		// Proceed with server shutdown
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}

		if s.cleanupFunc != nil {
			s.cleanupFunc()
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
