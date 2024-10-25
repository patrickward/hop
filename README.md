# Hop - Experimental

Simple HTTP helpers and middleware for my Go projects. This is essentially a collection of useful functions and
middleware that I've found myself reusing in multiple projects. It uses some 3rd party libraries, but I've tried to keep
the dependencies to a minimum.

- `form.go` - Form parsing routines that use
  the [github.com/go-playground/form/v4](https://github.com/go-playground/form)
  library.
- `json.go` - JSON parsing routines that use the standard library.
- `query.go` - Query parsing routines that use the standard library.
- `hopcheck` - A collection of validation functions for various data types.
- `hopconfig` - A simple configuration manager that uses the standard library's `encoding/json` package, env variables, and flags.
- `hoplogger` - A wrapper to set up a logger using the standard library's `slog` package. It also adds the [github.com/lmittmann/tint](https://github.com/lmittmann/tint) library for colorized output.
- `hopview` - A simple view manager for rendering HTML templates. It uses the [html/template](https://pkg.go.dev/html/template) package from the standard library.
- `hopware` - Middleware routines that mostly use the standard library. The PreventCSRF middleware uses the [github.com/justinas/nosurf](https://github.com/justinas/nosurf) library.

## TODO

- [ ] Add more useful tests
