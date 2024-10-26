# Hop - Experimental

`hop` is a light-weight web toolkit for Go that emphasizes simplicity and ease of use. It's my personal
collection of HTTP helpers and middleware that I've found useful in my projects. It helps me build web applications
without the complexity of larger frameworks, while attempting to stay close to the standard library.

Why the name `hop`? Because it’s a small, light-weight toolkit that helps me "hop" into web development quickly!

## Goals

This is primarily a personal toolkit that works well for my needs - it may or may not be your cup of tea. The goals
of the project are:

- Stay close to Go's standard library
- Keep dependencies minimal
- Use simple, predictable patterns
- Have just enough structure to get going quickly
- Be flexible and extensible

## Features

- `form.go` - Form parsing routines that use
  the [github.com/go-playground/form/v4](https://github.com/go-playground/form)
  library.
- `json.go` - JSON parsing routines that use the standard library.
- `query.go` - Query parsing routines that use the standard library.

## Sub-packages

- `auth` - (future) Authentication and authorization middleware.
- `cache` - (future) Caching middleware and helpers.
- `chain` - Middleware routines that mostly use the standard library.
- `check` - A collection of validation functions for various data types.
- `conf` - A simple configuration manager that uses the standard library's `encoding/json` package, env variables, and flags.
- `encode` - (future) A collection of encoding and decoding functions for various data types and formats (e.g. forms, json, etc).
- `keep` - (future) Session management interfaces, helpers, and middleware.
- `log` - A wrapper to set up a logger using the standard library's `slog` package.
- `mail` - (future) Email sending and receiving helpers.
- `render` - A simple view manager for rendering HTML templates. It uses the [html/template](https://pkg.go.dev/html/template) package from the standard library.
- `serve` - A simple HTTP server with basic routing and middleware support.

## TODO

- [ ] Add more useful tests
