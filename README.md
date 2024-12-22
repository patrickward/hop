# Hop - Experimental

⚠️ **EXPERIMENTAL**: This framework is under active development and the API changes frequently. Not recommended for production use unless you're willing to vendor the code.

Hop is an experimental, modular web application framework for Go, designed for building server-side rendered applications with HTMX integration.

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

```go
package main

import (
    "context"
    "log"

    "github.com/patrickward/hop"
    "github.com/patrickward/hop/conf"
    "github.com/patrickward/hop/render"
)

func main() {
    // Create app configuration
    cfg := &hop.AppConfig{
        Config: conf.NewConfig(),
        TemplateSources: render.Sources{
            "": embeddedTemplates,  // Your embedded templates
        },
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
