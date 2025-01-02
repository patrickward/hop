package route

import (
	"net/http"
	"path"
	"strings"
)

// Group represents a collection of routes with a common prefix and middleware
type Group struct {
	mux         *Mux
	prefix      string
	middleware  Chain
	parent      *Group // Track parent group for middleware inheritance
	independent bool   // If true, this group will not inherit middleware from parent
}

// Independent marks the group as independent, meaning it will not inherit middleware from the parent
func (g *Group) Independent() *Group {
	g.independent = true
	g.middleware = NewChain() // Clear middleware and start fresh
	return g
}

// HandleFunc registers a handler without method restrictions
func (g *Group) HandleFunc(pattern string, handler http.Handler) {
	g.handle(pattern, handler)
}

// Use registers middleware with the group
func (g *Group) Use(middleware ...Middleware) {
	g.middleware = g.middleware.Append(middleware...)
}

// Get registers a GET handler within the group
func (g *Group) Get(pattern string, handler http.Handler) {
	g.handle("GET "+pattern, handler)
}

// GetHandler registers a GET handler within the group with a handler that returns an error
func (g *Group) GetHandler(pattern string, handler http.Handler) {
	g.handle("GET "+pattern, handler)
}

// Post registers a POST handler within the group
func (g *Group) Post(pattern string, handler http.Handler) {
	g.handle("POST "+pattern, handler)
}

// Put registers a PUT handler within the group
func (g *Group) Put(pattern string, handler http.Handler) {
	g.handle("PUT "+pattern, handler)
}

// Delete registers a DELETE handler within the group
func (g *Group) Delete(pattern string, handler http.Handler) {
	g.handle("DELETE "+pattern, handler)
}

// Patch registers a PATCH handler within the group
func (g *Group) Patch(pattern string, handler http.Handler) {
	g.handle("PATCH "+pattern, handler)
}

// Options registers an OPTIONS handler within the group
func (g *Group) Options(pattern string, handler http.Handler) {
	g.handle("OPTIONS "+pattern, handler)
}

// Head registers a HEAD handler within the group
func (g *Group) Head(pattern string, handler http.Handler) {
	g.handle("HEAD "+pattern, handler)
}

// getMiddlewareChain returns all middleware in the chain from root to this group
func (g *Group) getMiddlewareChain() Chain {
	// If this group is independent, return only this group's middleware
	if g.independent {
		return g.middleware
	}

	// If this group is the root, combine mux middleware with this group's middleware
	if g.parent == nil {
		// Base case: combine mux middleware with this group's middleware
		return g.mux.middleware.Extend(g.middleware)
	}

	// Recursive case: get parent's chain and extend with this group's middleware
	return g.parent.getMiddlewareChain().Extend(g.middleware)
}

// handle registers a handler with the group's prefix and middleware chain
func (g *Group) handle(pattern string, handler http.Handler) {
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

	// Get the combined middleware chain based on independence
	var h http.Handler
	if g.independent {
		h = g.middleware.Then(handler)
	} else {
		// For non-independent groups, apply all middleware from outside in
		h = g.getMiddlewareChain().Then(handler)
	}

	// Register with parent mux
	g.mux.ServeMux.Handle(fullPattern, h)
}

// PrefixGroup creates a nested group with a common prefix and applies the provided group function
func (g *Group) PrefixGroup(prefix string, group GroupFunc) *Group {
	subGroup := &Group{
		mux:        g.mux,
		prefix:     path.Join(g.prefix, prefix),
		middleware: NewChain(),
		parent:     g, // Set this group as parent
	}

	if group != nil {
		group(subGroup)
	}

	return subGroup
}

// Group creates a nested group at the current prefix level and applies the provided group function
func (g *Group) Group(group GroupFunc) *Group {
	return g.PrefixGroup("", group)
}
