// Package hop provides an experimental, modular web application framework for Go.
package hop

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"

	"github.com/patrickward/hop/v2/alert"
	"github.com/patrickward/hop/v2/dispatch"
	"github.com/patrickward/hop/v2/render"
	"github.com/patrickward/hop/v2/serve"
	"github.com/patrickward/hop/v2/utils"
)

// OnTemplateDataFunc is a function type that takes an HTTP request and a pointer to a map of data.
// It represents a callback function that can be used to populate data for templates.
type OnTemplateDataFunc func(r *http.Request, data *map[string]any)

// AppConfig provides configuration options for creating a new App instance.
// It allows customization of core framework components including logging,
// template rendering, session management, and I/O configuration.
type AppConfig struct {
	// Environment is the application environment (default: "development")
	Environment string
	// Host is the host address to bind the server to (default: "" - all interfaces)
	Host string
	// Port is the port to bind the server to (default: 8080)
	Port int
	// Handler is the http.Handler to use for the server (default: http.DefaultServeMux)
	Handler http.Handler
	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled (default: 120s)
	IdleTimeout time.Duration
	// ReadTimeout is the maximum duration before timing out read of the request (default: 5s)
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out write of the response (default: 10s)
	WriteTimeout time.Duration
	// ShutdownTimeout is the maximum duration before timing out server shutdown (default: 10s)
	ShutdownTimeout time.Duration
	// Logger is the application's logging instance. If nil, a default logger will be created based on the configuration
	Logger *slog.Logger
	// TemplateFS is the file system to use for template files
	TemplateFS fs.FS
	// TemplateFuncs merges custom template functions into the default set of functions provided by hop. These are available in all templates.
	TemplateFuncs template.FuncMap
	// TemplateExt defines the extension for template files (default: ".html")
	TemplateExt string
	// TemplateLayoutsDir is the directory for layout templates (default: "layouts")
	TemplateLayoutsDir string
	// TemplatePartialsDir is the directory for partial templates (default: "partials")
	TemplatePartialsDir string
	// TemplatePagesDir is the directory for view templates (default: "pages")
	TemplatePagesDir string
	// TemplateErrorsDir is the directory for error templates (default: "errors")
	TemplateErrorsDir string
	// TemplateBaseLayout is the default layout to use for rendering templates (default: "base")
	TemplateBaseLayout string
	// TemplateErrorsLayout is the default layout to use for error templates (default: "base")
	TemplateErrorsLayout string
	// SessionStore provides the storage backend for sessions
	SessionStore scs.Store
	// SessionCookieLifetime is the duration for session cookies to persist (default: 168h)
	SessionLifetime time.Duration
	// SessionCookiePersist determines whether session cookies persist after the browser is closed (default: true)
	SessionCookiePersist bool
	// SessionCookieSameSite defines the SameSite attribute for session cookies (default: "lax")
	SessionCookieSameSite string
	// SessionCookieSecure determines whether session cookies are for secure connections only (default: true)
	SessionCookieSecure bool
	// SessionCookieHTTPOnly determines whether session cookies are HTTP-only (default: true)
	SessionCookieHTTPOnly bool
	// SessionCookiePath is the path for session cookies (default: "/")
	SessionCookiePath string
	// Stdout writer for standard output (default: os.Stdout)
	Stdout io.Writer
	// Stderr writer for error output (default: os.Stderr)
	Stderr io.Writer
}

// App represents the core application container that manages all framework components.
// It provides simple dependency injection, module management, and coordinates startup/shutdown
// of the application. It also ensures modules are started and stopped in the correct order.
type App struct {
	environment    string                      // environment (e.g. "production", "development", "test")
	host           string                      // host address
	port           int                         // port number
	flash          *alert.FlashManager         // flash message manager
	firstError     error                       // first error that occurred during initialization
	logger         *slog.Logger                // logger instance
	server         *serve.Server               // server instance
	handler        http.Handler                // the http.Handler for the server
	tm             *render.TemplateManager     // template manager instance
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
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Create events
	eventBus := dispatch.NewDispatcher(cfg.Logger)

	// Create the template manager
	var tm *render.TemplateManager
	if cfg.TemplateFS != nil {
		//return render.NewTemplateManager(templates.FS, render.WithExtension(".gtml"))
		tm = render.NewTemplateManager(
			cfg.TemplateFS,
			render.WithExtension(cfg.TemplateExt),
			render.WithFuncMap(cfg.TemplateFuncs),
			render.WithLayoutsDir(cfg.TemplateLayoutsDir),
			render.WithPartialsDir(cfg.TemplatePartialsDir),
			render.WithPagesDir(cfg.TemplatePagesDir),
			render.WithErrorsDir(cfg.TemplateErrorsDir),
			render.WithBaseLayout(cfg.TemplateBaseLayout),
			render.WithErrorsLayout(cfg.TemplateErrorsLayout),
		)
	}

	// Create the session manager
	sm := createSessionStore(&cfg)

	// Create router
	if cfg.Handler == nil {
		cfg.Handler = http.DefaultServeMux
	}

	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	// Create app
	app := &App{
		environment: cfg.Environment,
		flash:       alert.NewFlashManager("hop_flash", sm),
		host:        cfg.Host,
		port:        cfg.Port,
		logger:      cfg.Logger,
		events:      eventBus,
		modules:     make(map[string]Module),
		handler:     cfg.Handler,
		session:     sm,
		startOrder:  make([]string, 0),
		tm:          tm,
	}

	// Create server
	app.server = serve.NewServer(serve.Config{
		Address:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		IdleTimeout:     cfg.IdleTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
		Handler:         cfg.Handler,
		Logger:          cfg.Logger,
	})
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
		h.RegisterRoutes(a.handler)
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

// Flash returns the application's flash manager
func (a *App) Flash() *alert.FlashManager {
	return a.flash
}

// Handler returns the http.Handler for the app
func (a *App) Handler() http.Handler { return a.handler }

// Session returns the session manager instance for the app
func (a *App) Session() *scs.SessionManager { return a.session }

// TM returns the template manager instance for the app
func (a *App) TM() *render.TemplateManager { return a.tm }

// Host returns the host address for the app
func (a *App) Host() string { return a.host }

// Port returns the port number for the app
func (a *App) Port() int { return a.port }

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

	return render.NewResponse(a.tm).
		WithFlash(a.flash).
		Environment(a.environment).
		MergeData(a.NewTemplateData(r))
}

// NewTemplateData returns a map of data that can be used in a Go template, API response, etc.
// It includes the current user, environment, version, and other useful information.
func (a *App) NewTemplateData(r *http.Request) map[string]any {
	data := map[string]any{}

	// Add custom data from the onTemplateData callback. Set via app.OnTemplateData.
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

// ServerState returns the current state of the underlying server
func (a *App) ServerState() serve.ServerState {
	return a.server.State()
}

// IsServerRunning returns true if the server is running
func (a *App) IsServerRunning() bool {
	return a.server.IsRunning()
}

// -----------------------------------------------------------------------------
// Private functions
// -----------------------------------------------------------------------------

// createSessionStore creates a new session store based on the configuration
func createSessionStore(cfg *AppConfig) *scs.SessionManager {
	sameSite := utils.SameSiteFromString(cfg.SessionCookieSameSite)

	sessionMgr := scs.New()
	sessionMgr.Lifetime = cfg.SessionLifetime
	sessionMgr.Cookie.Persist = cfg.SessionCookiePersist
	sessionMgr.Cookie.SameSite = sameSite
	sessionMgr.Cookie.Secure = cfg.SessionCookieSecure
	sessionMgr.Cookie.HttpOnly = cfg.SessionCookieHTTPOnly
	sessionMgr.Cookie.Path = cfg.SessionCookiePath

	if cfg.SessionStore != nil {
		sessionMgr.Store = cfg.SessionStore
	}

	return sessionMgr
}
