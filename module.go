package hop

import (
	"context"
	"net/http"

	"github.com/patrickward/hop/route"
)

// Module represents a pluggable component that can be registered with the app
type Module interface {
	ID() string
	Init() error
}

// StartupModule represents a module that has a startup process
type StartupModule interface {
	Module
	Start(ctx context.Context) error
}

// ShutdownModule represents a module that has a shutdown process
type ShutdownModule interface {
	Module
	Stop(ctx context.Context) error
}

// HTTPModule represents a module that can add routes to the router
type HTTPModule interface {
	Module
	RegisterRoutes(router *route.Mux)
}

// EventModule represents a module that can handle events
type EventModule interface {
	Module
	RegisterEvents(events *EventBus)
}

// TemplateDataModule represents a module that can add data to the template context
type TemplateDataModule interface {
	Module
	OnTemplateData(r *http.Request, data *map[string]any)
}

// ConfigurableModule represents a module that can be configured
type ConfigurableModule interface {
	Module
	Configure(ctx context.Context, config any) error
}
