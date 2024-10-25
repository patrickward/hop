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

## Template Functions 

This package also provides a few template functions that can be used in HTML templates. 

Some of the functions include the following. See the `funcs.go` file for a complete list.

```html
{{/* Creating a map */}}
{{ $map := dict "name" "John" "age" 30 }}
{{/* Result: map[name:John age:30] */}}

{{/* Creating a slice */}}
{{ $items := slice "a" "b" "c" }}
{{/* Result: [a b c] */}}

{{/* Using key-value pairs */}}
{{ $user := kv "name" "John" }}
{{/* Result: map[name:John] */}}

{{/* String operations */}}
{{ "HELLO" | lower }}  
{{/* Result: hello */}}

{{ "hello" | upper }}  
{{/* Result: HELLO */}}

{{ "hello" | title }}  
{{/* Result: Hello */}}

{{ split "a,b,c" "," | join "-" }}  
{{/* Result: a-b-c */}}

{{/* Date formatting */}}
{{ $date := now | date "2006-01-02" }}  {{/* Result: current date in YYYY-MM-DD format */}}

{{/* Math operations */}}
{{ 5 | add 3 }}  {{/* Result: 8 */}}
{{ 10 | div 2 }}  {{/* Result: 5 */}}

{{/* Grouping items into 2 */}}
{{ $items := slice 1 2 3 4 5 6 }}
{{ group 2 $items }}  {{/* Result: [[1 2] [3 4] [5 6]] */}}
```

## TODO

- [ ] Add more useful tests
