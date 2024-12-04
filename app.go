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
	// core services
	logger *slog.Logger
	server *serve.Server
	router *route.Mux
	tm     *render.TemplateManager
	config *conf.Config
	events *EventBus

	modules     map[string]Module
	startOrder  []string
	dataModules []TemplateDataModule
	mu          sync.RWMutex

	onTemplateData OnTemplateDataFunc
}

// New creates a new application with core components
func New(cfg AppConfig) (*App, error) {
	// Create logger
	logger := createLogger(&cfg)

	// Create events
	eventBus := NewEventBus(logger)

	// Create template manager
	tm, err := createTemplateManager(&cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("error creating template manager: %w", err)
	}

	// Create session manager
	sm := createSessionStore(&cfg)

	// Create router
	router := route.New()

	// Create app
	app := &App{
		logger:     logger,
		events:     eventBus,
		router:     router,
		tm:         tm,
		config:     cfg.Config,
		modules:    make(map[string]Module),
		startOrder: make([]string, 0),
	}

	// Create server
	app.server = serve.NewServer(cfg.Config, logger, router, tm, sm)
	app.server.OnShutdown(func(ctx context.Context) error {
		return app.Stop(ctx)
	})

	return app, nil
}

// RegisterModule adds a module to the app
func (a *App) RegisterModule(m Module) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	id := m.ID()
	if _, exists := a.modules[id]; exists {
		return fmt.Errorf("module already registered: %s", id)
	}

	if err := m.Init(); err != nil {
		return fmt.Errorf("failed to initialize module %s: %s", id, err)
	}

	a.modules[id] = m
	a.startOrder = append(a.startOrder, id)

	if tdm, ok := m.(TemplateDataModule); ok {
		a.dataModules = append(a.dataModules, tdm)
	}

	if h, ok := m.(HTTPModule); ok {
		h.RegisterRoutes(a.router)
	}

	return nil
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
func (a *App) Events() *EventBus { return a.events }

// Router returns the router instance for the app
func (a *App) Router() *route.Mux { return a.router }

// Session returns the session manager instance for the app
func (a *App) Session() *scs.SessionManager { return a.server.Session() }

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

// NewResponse creates a new Response instance with the TemplateManager.
func (a *App) NewResponse(r *http.Request) (*render.Response, error) {
	if a.tm == nil {
		return nil, errors.New("template manager is not set")
	}

	resp := render.NewResponse(a.tm).Data(a.NewTemplateData(r))
	return resp, nil
}

// NewTemplateData returns a map of data that can be used in a Go template. It includes the current user, environment, version, and other useful information.
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

func createTemplateManager(cfg *AppConfig, logger *slog.Logger) (*render.TemplateManager, error) {
	return render.NewTemplateManager(
		cfg.TemplateSources,
		render.TemplateManagerOptions{
			Extension: cfg.TemplateExt,
			Funcs:     cfg.TemplateFuncs,
			Logger:    logger,
		})
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
	sameSite := http.SameSiteDefaultMode
	switch key {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteDefaultMode
	}
	return sameSite
}
