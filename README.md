# Hop - Experimental

⚠️ **EXPERIMENTAL**: This framework is under active development and the API changes frequently. Not recommended for
production use unless you're willing to vendor the code.

Hop is an experimental, modular web application framework for Go that makes it easy to build server-side rendered
applications with Go's `html/template` package and HTMX integration.

> This is more of a test bed for ideas than a production-ready framework. Use at your own risk.

## What is Hop?

Hop provides a simple foundation for building web applications that render HTML templates on the server. It handles
common web application concerns like:

- Template rendering with layouts, partials, and error pages
- Session management with configurable cookie settings
- Flash messages for user feedback
- Event dispatching between application components
- Modular architecture for organizing application features
- Built-in support for HTMX patterns

The framework is built around Go's standard `html/template` package and doesn't try to reinvent templating - it just
makes it easier to organize and render templates in a web application.

## Warning

This is not a general-purpose web framework. It was built for specific use cases and may not suit your needs. Consider
using established frameworks
like [Chi](https://github.com/go-chi/chi), [Echo](https://echo.labstack.com/), [Gin](https://gin-gonic.com/),
or [Fiber](https://gofiber.io/) for production applications.

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
"fmt"
"log"
"net/http"

    "github.com/patrickward/hop/v2"
)

func main() {
// Create a simple router
mux := http.NewServeMux()

    // Configure the application
    appCfg := hop.AppConfig{
        Environment:          "development",
        Host:                 "localhost",
        Port:                 8080,
        Handler:              mux,
        TemplateFS:           templates.FS, // embed.FS containing your templates
        TemplateFuncs:        funcs.NewFuncMap(), // custom template functions
        TemplateBaseLayout:   "base",
        TemplateErrorsLayout: "error", 
        SessionLifetime:      time.Hour * 24 * 7, // 1 week
        SessionCookieSecure:  false, // set to true in production
    }

    // Initialize the app
    app, err := hop.New(appCfg)
    if err != nil {
        log.Fatal(fmt.Errorf("error creating application: %w", err))
    }

    // Register a simple route that renders a template
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Create a response using the render package
        resp := app.NewResponse(r).
            Path("home").           // renders templates/pages/home.html
            Title("Welcome").
            Data("message", "Hello, World!")

        if err := resp.Write(w, r); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    // Start the server (this will block)
    if err := app.Start(context.Background()); err != nil {
        log.Fatal(fmt.Errorf("error starting server: %w", err))
    }
}
```

## Template Rendering

The `render` package provides a simple way to render templates through the `NewResponse` method on your app instance:

```go
func homeHandler(app *hop.App) http.HandlerFunc {
    return func (w http.ResponseWriter, r *http.Request) {
        // Create a response with template data
        resp := app.NewResponse(r).
        Path("home").   // template: pages/home.html
        Layout("base"). // layout: layouts/base.html -- base is the default layout, so this is optional
        Title("Home Page").
        Data("user", getCurrentUser(r)).
        Data("posts", getRecentPosts())

        // Write the response
        if err := resp.Write(w, r); err != nil {
            app.Logger().Error("template render error", "error", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        }
    }
}

// Register the handler
mux.HandleFunc("/", homeHandler(app))
```

## Template Organization and Structure

### Required Template Patterns

**Layout Templates (Required)**
- Layout templates **must** be defined with the `layout:` prefix: `{{define "layout:name"}}`
- Examples: `{{define "layout:base"}}`, `{{define "layout:minimal"}}`, `{{define "layout:admin"}}`
- This is the only hard requirement in the template system

### Recommended Conventions

**Directory Structure (Configurable)**

The default directory structure can be customized via `AppConfig`:
```

templates/
├── layouts/     # Layout templates (configurable: TemplateLayoutsDir)
├── pages/       # Page templates (configurable: TemplatePagesDir)  
├── partials/    # Reusable components (configurable: TemplatePartialsDir)
└── errors/      # Error page templates (configurable: TemplateErrorsDir)
```

**Template Naming (Your Choice)**

- **Page templates**: Name them however you like (e.g., `home.html`, `user/profile.html`)
- **Partials**: Common patterns include `@name` or `_name`, but any naming works
- **Error templates**: Typically named by status code (`404.html`, `500.html`)

**Template Definitions (Flexible)**

You can use any naming pattern for template definitions. Common patterns include:

- `{{define "page:main"}}` for page content and used within the layout
- `{{define "@header"}}` for partials. I like to use `@` for partials, but you can use `_` or any other prefix. 

### Example Layout Template
```
html
{{define "layout:base"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    {{template "@header" .}}
    {{template "page:main" .}}
    {{template "@footer" .}}
</body>
</html>
{{end}}
```

### Example Page Template

```html
{{define "page:main"}}
<div class="content">
    <h1>{{.Title}}</h1>
    <p>{{.Data.FooBar}}</p>
</div>
{{end}}
```

The key point is that **only the `layout:` prefix is mandatory**. Everything else - directory structure, file naming, and template definition patterns - is customizable based on your project's needs.

When referring to pages, you can use the `Path` method to specify the template path relative to the configured directories. For example, `app.NewResponse(r).Path("home")` will look for `templates/pages/home.html` by default. If no extension is provided, it will use the configured `TemplateExt` (default is `.html`).


## Configuration Options

The `AppConfig` struct provides the following options. See the `app.go` file for comments on each field.

```go
appCfg := hop.AppConfig{
    // Server settings
    Environment:         "production",
    Host:               "0.0.0.0",
    Port:               8080,
    Handler:            yourRouter,
    IdleTimeout:        time.Minute * 2,
    ReadTimeout:        time.Second * 5,
    WriteTimeout:       time.Second * 10,
    ShutdownTimeout:    time.Second * 10,

    // Template settings
    TemplateFS:              templates.FS,
    TemplateFuncs:           customFuncs,
    TemplateExt:             ".html",
    TemplateLayoutsDir:      "layouts",
    TemplatePartialsDir:     "partials", 
    TemplatePagesDir:        "pages",
    TemplateErrorsDir:       "errors",
    TemplateBaseLayout:      "base",
    TemplateErrorsLayout:    "error", // new option replacing SetSystemPagesLayout

    // Session settings
    SessionStore:            redisStore,
    SessionLifetime:         time.Hour * 24 * 7,
    SessionCookiePersist:    true,
    SessionCookieSameSite:   "lax",
    SessionCookieSecure:     true,
    SessionCookieHTTPOnly:   true,
    SessionCookiePath:       "/",

    // Logging
    Logger:                  slog.New(slog.NewJSONHandler(os.Stdout, nil)),
}
```

## Creating a Module

Modules let you organize related functionality:

```go
type MyModule struct {
    logger *slog.Logger
}

func NewMyModule(logger *slog.Logger) *MyModule {
    return &MyModule{logger: logger}
}

func (m *MyModule) ID() string {
    return "mymodule"
}

func (m *MyModule) Init() error {
    m.logger.Info("MyModule initialized")
    return nil
}

// Optional: implement StartupModule if you need startup logic
func (m *MyModule) Start(ctx context.Context) error {
    m.logger.Info("MyModule started")
    return nil
}

// Optional: implement ShutdownModule if you need cleanup logic
func (m *MyModule) Stop(ctx context.Context) error {
    m.logger.Info("MyModule stopped")
    return nil
}

// Register the module
app.RegisterModule(NewMyModule(app.Logger()))
```

## Flash Messages and Form Handling

```go
func createPostHandler(app *hop.App) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            // Process form submission
            if err := processForm(r); err != nil {
                // Show error and re-render form
                resp := app.NewResponse(r).
                    Path("posts/create").
                    FieldError("title", "Title is required").
                    Message(alert.TypeError, "Please fix the errors below")
                
                resp.Write(w, r)
                return
            }

            // Success - add flash message and redirect
            app.Flash().Success(r.Context(), "Post created successfully!")
            http.Redirect(w, r, "/posts", http.StatusSeeOther)
            return
        }

        // Show create form
        resp := app.NewResponse(r).
            Path("posts/create").
            Title("Create Post")
        
        resp.Write(w, r)
    }
}
```
