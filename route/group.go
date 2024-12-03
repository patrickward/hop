package route

import (
	"net/http"
	"path"
	"strings"
)

// Group represents a collection of routes with a common prefix and middleware
type Group struct {
	mux        *Mux
	prefix     string
	middleware Chain
	parent     *Group // Track parent group for middleware inheritance
}

// HandleFunc registers a handler without method restrictions
func (g *Group) HandleFunc(pattern string, handler http.HandlerFunc) {
	g.handle(pattern, handler)
}

// Use registers middleware with the group
func (g *Group) Use(middleware ...Middleware) {
	g.middleware = g.middleware.Append(middleware...)
}

// Get registers a GET handler within the group
func (g *Group) Get(pattern string, handler http.HandlerFunc) {
	g.handle("GET "+pattern, handler)
}

// Post registers a POST handler within the group
func (g *Group) Post(pattern string, handler http.HandlerFunc) {
	g.handle("POST "+pattern, handler)
}

// Put registers a PUT handler within the group
func (g *Group) Put(pattern string, handler http.HandlerFunc) {
	g.handle("PUT "+pattern, handler)
}

// Delete registers a DELETE handler within the group
func (g *Group) Delete(pattern string, handler http.HandlerFunc) {
	g.handle("DELETE "+pattern, handler)
}

// Patch registers a PATCH handler within the group
func (g *Group) Patch(pattern string, handler http.HandlerFunc) {
	g.handle("PATCH "+pattern, handler)
}

// Options registers an OPTIONS handler within the group
func (g *Group) Options(pattern string, handler http.HandlerFunc) {
	g.handle("OPTIONS "+pattern, handler)
}

// Head registers a HEAD handler within the group
func (g *Group) Head(pattern string, handler http.HandlerFunc) {
	g.handle("HEAD "+pattern, handler)
}

// getMiddlewareChain returns all middleware in the chain from root to this group
func (g *Group) getMiddlewareChain() Chain {
	if g.parent == nil {
		// Base case: combine mux middleware with this group's middleware
		return g.mux.middleware.Extend(g.middleware)
	}

	// Recursive case: get parent's chain and extend with this group's middleware
	return g.parent.getMiddlewareChain().Extend(g.middleware)
}

// handle registers a handler with the group's prefix and middleware chain
func (g *Group) handle(pattern string, handler http.HandlerFunc) {
	// Extract method if present
	var method string
	if len(pattern) > 0 && pattern[0] != '/' {
		parts := strings.SplitN(pattern, " ", 2)
		if len(parts) == 2 {
			method = parts[0]
			pattern = parts[1]
		}
	}

	// Combine group prefix with pattern
	fullPattern := path.Join(g.prefix, pattern)

	if method != "" {
		// Register the route with the registry
		g.mux.registry.register(fullPattern, method)
		// Prepend method to pattern for mux registration
		fullPattern = method + " " + fullPattern
	}

	// Get the combined middleware chain
	chain := g.getMiddlewareChain()

	// Apply all middleware from outside in
	h := chain.ThenFunc(handler)

	// Register with parent mux
	g.mux.ServeMux.Handle(fullPattern, h)
}

// Group creates a nested group
func (g *Group) Group(prefix string, middleware ...Middleware) *Group {
	return &Group{
		mux:        g.mux,
		prefix:     path.Join(g.prefix, prefix),
		middleware: NewChain(middleware...),
		parent:     g, // Set this group as parent
	}
}
