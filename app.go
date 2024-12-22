/*
Package hop provides an experimental, modular web application framework for Go.

# ⚠️ Important Notice

This framework is in active development and is currently EXPERIMENTAL.

  - The API is unstable and changes frequently without notice
  - Documentation may be incomplete or outdated
  - Not recommended for production use unless you're willing to vendor the code
  - Built primarily for specific use cases and may not be suitable for general use
  - Limited community support and testing

Consider using established frameworks like Chi, Echo, Gin, or Fiber for production applications.
If you decide to use Hop, be prepared to:
  - Handle breaking changes regularly
  - Read and understand the source code
  - Potentially fork and maintain your own version
  - Contribute fixes and improvements back to the project

# What is Hop?

Hop is designed to be a flexible, maintainable web framework that follows Go idioms
and best practices. It provides a modular architecture where functionality can be
easily extended through a plugin system while maintaining a clean separation of concerns.

# Core Features

  - Modular architecture with pluggable components
  - Built-in template rendering with layouts and partials
  - Session management
  - Configuration management
  - Structured logging
  - Event dispatching
  - HTTP routing with middleware support
  - Background task management
  - Graceful shutdown handling

# Getting Started

	import "github.com/patrickward/hop"

	func main() {
		// Create app configuration
		appConfig := hop.AppConfig{
			Config:          &cfg.Hop,
			Logger:          logger,
			TemplateSources: render.Sources{"-": &templates.Files},
			TemplateFuncs:   funcs.NewTemplateFuncMap(authorizer),
			SessionStore:    store,
			Stdout:          stdout,
			Stderr:          stderr,
		}

		// Initialize the app
		app, err := hop.New(cfg)
		if err != nil {
			log.Fatal(err)
		}

		// Register modules
		app.RegisterModule(mymodule.New())

		// Start the app
		if err := app.Start(context.Background()); err != nil {
			log.Fatal(err)
		}
	}

# Modules

Hop uses a module system to organize and extend functionality. Modules are Go types that
implement one or more of the following interfaces:

  - Module (required): Base interface for all modules
  - StartupModule: For modules that need initialization at startup
  - ShutdownModule: For modules that need cleanup at shutdown
  - HTTPModule: For modules that provide HTTP routes
  - DispatcherModule: For modules that handle events
  - TemplateDataModule: For modules that provide template data
  - ConfigurableModule: For modules that require configuration

Creating a basic module:

	```go
	type MyModule struct {}

	func (m *MyModule) ID() string {
	    return "mymodule"
	}

	func (m *MyModule) Init() error {
	    return nil
	}
	```

# Template Rendering

Hop includes a template system that supports:

  - Multiple template sources
  - Layout templates
  - Partial templates
  - Custom template functions
  - Automatic template caching
  - HTMX integration

Example template usage:

	```go
	resp := app.NewResponse(r).
	    Layout("main").
	    Path("pages/home").
	    WithData(map[string]any{
	        "Title": "Welcome",
	    }).
	    StatusOK()
	resp.Render(w, r)
	```

# Configuration

The framework uses a structured configuration system that can be customized:

  - Configuration file support
  - Environment variable overrides
  - Type-safe configuration access

# Event System

Hop includes an event dispatcher for loose coupling between components:

  - Publish/subscribe pattern
  - Async event handling
  - Type-safe event definitions
  - Module-specific event handlers

The event system is useful for small, monolithic apps or direct interaction between modules within a single server.

# Best Practices

When building applications with Hop:

1. Organize related functionality into modules
2. Use dependency injection via the App container
3. Handle graceful shutdown in modules that need cleanup
4. Use the event system for cross-module communication
5. Implement appropriate interfaces based on module needs
6. Use structured logging for better observability
7. Follow Go idioms and conventions

For more information, see the documentation for individual types and interfaces.
*/
package hop

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"

	"github.com/patrickward/hop/conf"
	"github.com/patrickward/hop/dispatch"
	"github.com/patrickward/hop/log"
	"github.com/patrickward/hop/render"
	"github.com/patrickward/hop/render/htmx"
	"github.com/patrickward/hop/route"
	"github.com/patrickward/hop/serve"
	"github.com/patrickward/hop/utils"
)

// OnTemplateDataFunc is a function type that takes an HTTP request and a pointer to a map of data.
// It represents a callback function that can be used to populate data for templates.
type OnTemplateDataFunc func(r *http.Request, data *map[string]any)

// AppConfig provides configuration options for creating a new App instance.
// It allows customization of core framework components including logging,
// template rendering, session management, and I/O configuration.
type AppConfig struct {
	// Config holds the application's configuration settings
	Config *conf.HopConfig
	// Logger is the application's logging instance. If nil, a default logger will be created based on the configuration
	Logger *slog.Logger
	// TemplateSources defines the sources for template files. Multiple sources can be provided with different prefixes
	TemplateSources render.Sources
	// TemplateFuncs merges custom template functions into the default set of functions provided by hop. These are available in all templates.
	TemplateFuncs template.FuncMap
	// TemplateExt defines the extension for template files (default: ".html")
	TemplateExt string
	// SessionStore provides the storage backend for sessions
	SessionStore scs.Store
	// Stdout writer for standard output (default: os.Stdout)
	Stdout io.Writer
	// Stderr writer for error output (default: os.Stderr)
	Stderr io.Writer
}

// App represents the core application container that manages all framework components.
// It provides simple dependency injection, module management, and coordinates startup/shutdown
// of the application. App implements graceful shutdown and ensures modules are started
// and stopped in the correct order.
type App struct {
	firstError     error                       // first error that occurred during initialization
	logger         *slog.Logger                // logger instance
	server         *serve.Server               // server instance
	router         *route.Mux                  // router instance
	tm             *render.TemplateManager     // template manager instance
	config         *conf.HopConfig             // configuration
	events         *dispatch.Dispatcher        // event bus instance
	session        *scs.SessionManager         // session manager instance
	modules        map[string]Module           // map of modules by ID
	startOrder     []string                    // order in which modules should be started / stopped in reverse
	dataModules    []TemplateDataModule        // modules that provide template data
	mu             sync.RWMutex                // mutex for modules map
	onTemplateData OnTemplateDataFunc          // callback function for populating template data
	onShutdown     func(context.Context) error // callback function for shutting down the app. This is called when the server is shutting down.
}

// New creates a new application with core components
func New(cfg AppConfig) (*App, error) {
	// Create logger
	logger := createLogger(&cfg)

	// Create events
	eventBus := dispatch.NewDispatcher(logger)

	// Create template manager
	var tm *render.TemplateManager
	if len(cfg.TemplateSources) > 0 {
		var err error
		tm, err = render.NewTemplateManager(
			cfg.TemplateSources,
			render.TemplateManagerOptions{
				Extension: cfg.TemplateExt,
				Funcs:     cfg.TemplateFuncs,
				Logger:    logger,
			})
		if err != nil {
			return nil, fmt.Errorf("error creating template manager: %w", err)
		}

	}

	// Create session manager
	sm := createSessionStore(&cfg)

	// Create router
	router := route.New()

	// Create app
	app := &App{
		config:     cfg.Config,
		logger:     logger,
		events:     eventBus,
		modules:    make(map[string]Module),
		router:     router,
		session:    sm,
		startOrder: make([]string, 0),
		tm:         tm,
	}

	// Create server
	app.server = serve.NewServer(cfg.Config, logger, router)
	app.server.OnShutdown(func(ctx context.Context) error {
		return app.Stop(ctx)
	})

	return app, nil
}

// -----------------------------------------------------------------------------

// Error returns the first error that occurred during initialization
func (a *App) Error() error {
	return a.firstError
}

// RegisterModule adds a module to the app
func (a *App) RegisterModule(m Module) *App {
	a.mu.Lock()
	defer a.mu.Unlock()

	// If we already have an error, don't bother trying to register more modules
	if a.firstError != nil {
		return a
	}

	id := m.ID()
	if _, exists := a.modules[id]; exists {
		a.firstError = fmt.Errorf("module already registered: %s", id)
		return a
	}

	if err := m.Init(); err != nil {
		a.firstError = fmt.Errorf("failed to initialize module %s: %s", id, err)
		return a
	}

	a.modules[id] = m
	a.startOrder = append(a.startOrder, id)

	if tdm, ok := m.(TemplateDataModule); ok {
		a.dataModules = append(a.dataModules, tdm)
	}

	if h, ok := m.(HTTPModule); ok {
		h.RegisterRoutes(a.router)
	}

	return a
}

// GetModule returns a module by ID
func (a *App) GetModule(id string) (Module, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	m, exists := a.modules[id]
	if !exists {
		return nil, fmt.Errorf("module not found: %s", id)
	}

	return m, nil
}

// StartModules initializes and starts all modules without starting the server
func (a *App) StartModules(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var errs []error

	// Start modules that implement StartupModule
	for _, id := range a.startOrder {
		m := a.modules[id]
		if s, ok := m.(StartupModule); ok {
			a.logger.Info("starting module", slog.String("module", id))
			if err := s.Start(ctx); err != nil {
				errs = append(errs, err)
				a.logger.Error("failed to start module %s: %w", id, err)
			}
		}
	}

	return errors.Join(errs...)
}

// Start initializes the app and starts all modules and the server
func (a *App) Start(ctx context.Context) error {
	// First start all modules
	if err := a.StartModules(ctx); err != nil {
		return err
	}

	// Then start the server (this will block)
	if err := a.server.Start(); err != nil {
		a.logger.Error("failed to start server", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// ShutdownServer gracefully shuts down the server
func (a *App) ShutdownServer(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// Stop gracefully shuts down the app and all modules. This is only called when the server is shutting down.
func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("shutting down app")
	a.mu.RLock()
	defer a.mu.RUnlock()

	var errs []error

	// Stop modules in reverse order that implement ShutdownModule
	for i := len(a.startOrder) - 1; i >= 0; i-- {
		id := a.startOrder[i]
		m := a.modules[id]
		if sm, ok := m.(ShutdownModule); ok {
			a.logger.Info("stopping module", "module", id)
			if err := sm.Stop(ctx); err != nil {
				errs = append(errs, err)
				a.logger.Error("sailed to stop module", slog.String("module", id), slog.String("error", err.Error()))
			}
		}
	}

	// run the onShutdown callback if it's set
	if a.onShutdown != nil {
		if err := a.onShutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Logger returns the logger instance for the app
func (a *App) Logger() *slog.Logger { return a.logger }

// Dispatcher returns the event bus for the app
func (a *App) Dispatcher() *dispatch.Dispatcher { return a.events }

// Router returns the router instance for the app
func (a *App) Router() *route.Mux { return a.router }

// Session returns the session manager instance for the app
func (a *App) Session() *scs.SessionManager { return a.session }

// TM returns the template manager instance for the app
func (a *App) TM() *render.TemplateManager { return a.tm }

// Config returns the configuration for the app
func (a *App) Config() *conf.HopConfig { return a.config }

// RunInBackground runs a function in the background via the server
func (a *App) RunInBackground(r *http.Request, fn func() error) {
	a.server.BackgroundTask(r, fn)
}

// OnTemplateData registers a function that populates template data each time a template is rendered.
func (a *App) OnTemplateData(fn OnTemplateDataFunc) {
	a.onTemplateData = fn
}

// OnShutdown registers a function to be called when the app is shutting down
func (a *App) OnShutdown(fn func(context.Context) error) {
	a.onShutdown = fn
}

// NewResponse creates a new Response instance with the TemplateManager.
func (a *App) NewResponse(r *http.Request) *render.Response {
	if a.tm == nil {
		panic("template manager not initialized - this app does not support rendering templates")
	}

	return render.NewResponse(a.tm).WithData(a.NewTemplateData(r))
}

// NewTemplateData returns a map of data that can be used in a Go template, API response, etc.
// It includes the current user, environment, version, and other useful information.
func (a *App) NewTemplateData(r *http.Request) map[string]any {
	cacheBuster := func() string {
		return time.Now().UTC().Format("20060102150405")
	}

	data := map[string]any{
		//"CurrentUser":        auth.GetCurrentUserFromContext(r),
		"Environment":        a.config.App.Environment,
		"IsDevelopment":      a.config.App.Environment == "development",
		"IsProduction":       a.config.App.Environment == "production",
		"CSRFToken":          nosurf.Token(r),
		"BaseURL":            a.config.Server.BaseURL,
		"CacheBuster":        cacheBuster,
		"RequestPath":        r.URL.Path,
		"IsHome":             r.URL.Path == "/",
		"IsHTMXRequest":      htmx.IsHtmxRequest(r),
		"IsBoostedRequest":   htmx.IsBoostedRequest(r),
		"IsAnyHtmxRequest":   htmx.IsAnyHtmxRequest(r),
		"MaintenanceEnabled": a.config.Maintenance.Enabled,
		"MaintenanceMessage": a.config.Maintenance.Message,
	}

	// Add custom data from the callback function
	if a.onTemplateData != nil {
		newData := make(map[string]any)
		a.onTemplateData(r, &newData)
		utils.DeepMerge(&data, newData)
	}

	// Allow modules that are of type TemplateDataModule to contribute data
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, tdm := range a.dataModules {
		moduleData := make(map[string]any)
		tdm.OnTemplateData(r, &moduleData)
		utils.DeepMerge(&data, moduleData)
	}

	return data
}

// -----------------------------------------------------------------------------
// Route Functions (TEMPORARY)
// -----------------------------------------------------------------------------

// AddRoute adds a new route to the server, using the newer v1.22 http.Handler interface. It takes a pattern, an http.Handler, and an optional list of middleware.
func (a *App) AddRoute(pattern string, handler http.Handler, middleware ...route.Middleware) {
	if len(middleware) > 0 {
		// Create a chain of middleware and wrap the handler
		chain := route.NewChain(middleware...).Then(handler)
		a.router.Handle(pattern, chain)
		return
	}
	a.router.Handle(pattern, handler)
}

// AddChainedRoute adds a new route to the server with a chain of middleware
// It takes a pattern, an http.Handler, and a route.Chain struct
func (a *App) AddChainedRoute(pattern string, handler http.Handler, chain route.Chain) {
	a.router.Handle(pattern, chain.Then(handler))
}

// AddRoutes adds multiple routes to the server. It takes a map of patterns to http.Handlers and an optional list of middleware.
func (a *App) AddRoutes(routes map[string]http.Handler, middleware ...route.Middleware) {
	for pattern, handler := range routes {
		if len(middleware) > 0 {
			a.AddRoute(pattern, handler, middleware...)
			continue
		}
		a.AddRoute(pattern, handler)
	}
}

// AddChainedRoutes adds multiple routes to the server with a chain of middleware
func (a *App) AddChainedRoutes(routes map[string]http.Handler, chain route.Chain) {
	for pattern, handler := range routes {
		a.AddChainedRoute(pattern, handler, chain)
	}
}

// -----------------------------------------------------------------------------
// Private functions
// -----------------------------------------------------------------------------

func createLogger(cfg *AppConfig) *slog.Logger {
	if cfg.Logger == nil {
		if cfg.Stderr == nil {
			cfg.Stderr = os.Stderr
		}
		logger := log.NewLogger(log.Options{
			Format:      cfg.Config.Log.Format,
			IncludeTime: cfg.Config.Log.IncludeTime,
			Level:       cfg.Config.Log.Level,
			Verbose:     cfg.Config.Log.Verbose,
			Writer:      cfg.Stderr,
		})
		cfg.Logger = logger
	}

	return cfg.Logger
}

// createSessionStore creates a new session store based on the configuration
// TODO: Add support for other session stores
func createSessionStore(cfg *AppConfig) *scs.SessionManager {
	sameSite := utils.SameSiteFromString(cfg.Config.Session.CookieSameSite)

	sessionMgr := scs.New()
	sessionMgr.Lifetime = cfg.Config.Session.Lifetime.Duration
	sessionMgr.Cookie.Persist = cfg.Config.Session.CookiePersist
	sessionMgr.Cookie.SameSite = sameSite
	sessionMgr.Cookie.Secure = cfg.Config.Session.CookieSecure
	sessionMgr.Cookie.HttpOnly = cfg.Config.Session.CookieHTTPOnly
	sessionMgr.Cookie.Path = cfg.Config.Session.CookiePath

	if cfg.SessionStore != nil {
		sessionMgr.Store = cfg.SessionStore
	}

	return sessionMgr
}
