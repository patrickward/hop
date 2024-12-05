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
	"github.com/patrickward/hop/events"
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

// AppConfig is a configuration for the app
type AppConfig struct {
	Config          *conf.Config
	Logger          *slog.Logger
	TemplateSources render.Sources
	TemplateFuncs   template.FuncMap
	TemplateExt     string
	SessionStore    scs.Store
	Stdout          io.Writer
	Stderr          io.Writer
}

// App holds core components that most services need
type App struct {
	firstError     error                   // first error that occurred during initialization
	logger         *slog.Logger            // logger instance
	server         *serve.Server           // server instance
	router         *route.Mux              // router instance
	tm             *render.TemplateManager // template manager instance
	config         *conf.Config            // configuration
	events         *events.Bus             // event bus instance
	session        *scs.SessionManager     // session manager instance
	modules        map[string]Module       // map of modules by ID
	startOrder     []string                // order in which modules should be started / stopped in reverse
	dataModules    []TemplateDataModule    // modules that provide template data
	mu             sync.RWMutex            // mutex for modules map
	onTemplateData OnTemplateDataFunc      // callback function for populating template data
}

// New creates a new application with core components
func New(cfg AppConfig) (*App, error) {
	// Create logger
	logger := createLogger(&cfg)

	// Create events
	eventBus := events.NewEventBus(logger)

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

// Start initializes the app and starts all modules
func (a *App) Start(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var errs []error

	// Start modules that implement StartupModule
	for _, id := range a.startOrder {
		m := a.modules[id]
		if s, ok := m.(StartupModule); ok {
			a.logger.Info("Starting module", slog.String("module", id))
			if err := s.Start(ctx); err != nil {
				errs = append(errs, err)
				a.logger.Error("failed to start module %s: %w", id, err)
			}
		}
	}

	// Start the server
	if err := a.server.Start(); err != nil {
		errs = append(errs, err)
		a.logger.Error("failed to start server", slog.String("error", err.Error()))
	}

	return errors.Join(errs...)
}

// Stop gracefully shuts down the app and all modules
func (a *App) Stop(ctx context.Context) error {
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
				a.logger.Error("failed to stop module", slog.String("module", id), slog.String("error", err.Error()))
			}
		}
	}

	return errors.Join(errs...)
}

// Logger returns the logger instance for the app
func (a *App) Logger() *slog.Logger { return a.logger }

// Events returns the event bus for the app
func (a *App) Events() *events.Bus { return a.events }

// Router returns the router instance for the app
func (a *App) Router() *route.Mux { return a.router }

// Session returns the session manager instance for the app
func (a *App) Session() *scs.SessionManager { return a.session }

// TM returns the template manager instance for the app
func (a *App) TM() *render.TemplateManager { return a.tm }

// Config returns the configuration for the app
func (a *App) Config() *conf.Config { return a.config }

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
	a.server.OnShutdown(fn)
}

// NewResponse creates a new Response instance with the TemplateManager.
func (a *App) NewResponse(r *http.Request) (*render.Response, error) {
	if a.tm == nil {
		return nil, fmt.Errorf("template manager not initialized - this app does not support rendering templates")
	}

	resp := render.NewResponse(a.tm).Data(a.NewTemplateData(r))
	return resp, nil
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
		"Company":            a.config.Company,
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
	sameSite := sameSiteFromString(cfg.Config.Session.CookieSameSite)

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

func sameSiteFromString(key string) http.SameSite {
	switch key {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
