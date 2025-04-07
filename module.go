package hop

import (
	"context"
	"net/http"

	"github.com/patrickward/hop/v2/dispatch"
)

// Module is the base interface that all modules must implement. It provides
// identification and initialization capabilities for the module system.
type Module interface {
	// ID returns a unique identifier for the module
	ID() string

	// Init performs any necessary module initialization
	// It is called when the module is registered with the application
	Init() error
}

// StartupModule is implemented by modules that need to perform actions
// during application startup. The Start method will be called after all
// modules are initialized but before the HTTP server begins accepting
// connections.
type StartupModule interface {
	Module
	// Start performs startup actions for the module
	// The provided context will be canceled when the application begins shutdown
	Start(ctx context.Context) error
}

// ShutdownModule is implemented by modules that need to perform cleanup
// actions during application shutdown. Modules are shut down in reverse
// order of their startup.
type ShutdownModule interface {
	Module
	// Stop performs cleanup actions for the module
	// It should respect the provided context's deadline for graceful shutdown
	Stop(ctx context.Context) error
}

// HTTPModule is implemented by modules that provide HTTP routes.
// The RegisterRoutes method is called after module initialization
// to set up any routes the module provides.
type HTTPModule interface {
	Module
	// RegisterRoutes adds the module's routes to the provided router
	RegisterRoutes(handler http.Handler)
}

// DispatcherModule is implemented by modules that handle application events.
// The RegisterEvents method is called after initialization to set up any
// event handlers the module provides.
type DispatcherModule interface {
	Module
	// RegisterEvents registers the module's event handlers with the dispatcher
	RegisterEvents(events *dispatch.Dispatcher)
}

// TemplateDataModule is implemented by modules that provide data to templates.
// The OnTemplateData method is called for each template render to allow
// the module to add its data to the template context.
type TemplateDataModule interface {
	Module
	// OnTemplateData allows the module to add data to the template context
	// The provided map will be merged with other template data
	OnTemplateData(r *http.Request, data *map[string]any)
}

// ConfigurableModule is implemented by modules that require configuration
// beyond basic initialization. The Configure method is called after Init
// but before Start.
type ConfigurableModule interface {
	Module
	// Configure applies the provided configuration to the module
	// The config parameter should be type asserted to the module's
	// specific configuration type
	Configure(ctx context.Context, config any) error
}
