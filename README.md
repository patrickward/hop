# HyperCore - Experimental

Simple HTTP helpers and middleware for my Go projects. This is essentially a collection of useful functions and
middleware that I've found myself reusing in multiple projects. It uses some 3rd party libraries, but I've tried to keep
the dependencies to a minimum.

- `form` - Form parsing routines that use the [github.com/go-playground/form/v4](https://github.com/go-playground/form)
  library.
- `json` - JSON parsing routines that use the standard library.
- `middleware` - Middleware routines that mostly use the standard library. The PreventCSRF middleware uses
  the [github.com/justinas/nosurf](https://github.com/justinas/nosurf) library.
- `query` - Query parsing routines that use the standard library.
- `slogger` - A wrapper to set up a logger using the standard library's `slog` package. It also adds
  the [github.com/lmittmann/tint](https://github.com/lmittmann/tint) library for colorized output.

## TODO

- [ ] Add more useful tests
