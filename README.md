# Hop - Experimental

⚠️ **EXPERIMENTAL**: This framework is under active development and the API changes frequently. Not recommended for production use unless you're willing to vendor the code.

Hop is an experimental, modular web application framework for Go, designed for building server-side rendered applications with HTMX integration.

> This is more of a test bed for ideas than a production-ready framework. Use at your own risk.

## Warning

This is not a general-purpose web framework. It was built for specific use cases and may not suit your needs. Consider using established frameworks like [Chi](https://github.com/go-chi/chi), [Echo](https://echo.labstack.com/), [Gin](https://gin-gonic.com/), or [Fiber](https://gofiber.io/) for production applications.

## Features

- 🧩 Modular architecture with plugin system
- 📝 Template rendering with layouts and partials
- 🔄 Built-in HTMX support
- 📦 Session management
- 🎯 Event dispatching
- 🛣️ HTTP routing with middleware
- 📋 Configuration management
- 📝 Structured logging

## Quick Start

TBD

```go
package main

import (
    "context"
    "log"

    "github.com/patrickward/hop/v2"
	
   "github.com/<owner>/<project>/internal/app"     
)

func main() {
	// Load configuration from somewhere
	cfg, err := app.NewConfigFromEnv(envPrefix)
	
	// Create app
	appCfg := hop.AppConfig{
		Environment:           cfg.Environment,
		Host:                  cfg.ServerHost,
		Port:                  cfg.ServerPort,
		Handler:               mux,
		IdleTimeout:           cfg.ServerIdleTimeout,
		ReadTimeout:           cfg.ServerReadTimeout,
		WriteTimeout:          cfg.ServerWriteTimeout,
		ShutdownTimeout:       cfg.ServerShutdownTimeout,
		Logger:                logger,
		SessionStore:          store,
		SessionLifetime:       cfg.SessionLifetime,
		SessionCookiePersist:  cfg.SessionCookiePersist,
		SessionCookieSameSite: cfg.SessionCookieSameSite,
		SessionCookieSecure:   cfg.SessionCookieSecure,
		SessionCookieHTTPOnly: cfg.SessionCookieHTTPOnly,
		SessionCookiePath:     cfg.SessionCookiePath,
		TemplateFS:            templates.FS,
		TemplateFuncs:         funcs.NewFuncMap(),
		Stdout:                stdout,
		Stderr:                stderr,
	}


	// Initialize the app
	hopApp, err := hop.New(appConfig)
	if err != nil {
		return fmt.Errorf("error creating application: %w", err)
	}

	hopApp.SetSystemPagesLayout("minimal")
	
	// Set up the app
	// ===============
	a := app.New(app.Settings{
		App:        hopApp,
		Config:     cfg,
		DB:         d,
		Logger:     logger,
		Mailer:     mailer,
		Name:       AppName,
		Version:    AppVersion,
		BuildTime:  BuildTime,
		CommitHash: CommitHash,
	})
	

	// Start the server, blocking until it exits
	// ==============
	if err := a.Start(context.Background()); err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

}
```

## Creating a Module

```go
type MyModule struct{}

func New() *MyModule {
    return &MyModule{}
}

func (m *MyModule) ID() string {
    return "mymodule"
}

func (m *MyModule) Init() error {
    return nil
}
```
