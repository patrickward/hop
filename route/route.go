package route

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

// Mux extends http.ServeMux with additional routing features.
// It also provides a middleware chain for adding middleware to routes.
type Mux struct {
	*http.ServeMux
	middleware      Chain
	registry        *routeRegistry
	notFoundHandler http.HandlerFunc
}

// New creates a new Mux instance
func New(middleware ...Middleware) *Mux {
	mux := &Mux{
		ServeMux:   http.NewServeMux(),
		middleware: NewChain(middleware...),
		registry:   newRouteRegistry(),
	}

	// Register the default route to handle OPTIONS and NotFound
	mux.ServeMux.HandleFunc("/", mux.handleOptions)

	return mux
}

// Use adds middleware to the Mux
func (m *Mux) Use(middleware ...Middleware) {
	m.middleware = m.middleware.Append(middleware...)
}

// Group creates a new route group with the given prefix and middleware
func (m *Mux) Group(prefix string, middleware ...Middleware) *Group {
	return &Group{
		mux:        m,
		prefix:     prefix,
		middleware: NewChain(middleware...),
		parent:     nil, // Root group has no parent
	}
}

// Home registers a handler for the root path
func (m *Mux) Home(handler http.HandlerFunc) {
	m.HandleFunc("/{$}", handler)
}

// NotFound registers a handler for when no routes match
func (m *Mux) NotFound(handler http.HandlerFunc) {
	m.notFoundHandler = handler
}

// handle registers a handler with middleware
func (m *Mux) handle(pattern string, handler http.HandlerFunc) {
	// Extract method if present
	var method string
	if len(pattern) > 0 && pattern[0] != '/' {
		parts := strings.SplitN(pattern, " ", 2)
		if len(parts) == 2 {
			method = parts[0]
			pattern = parts[1]
		}
	}

	// Register the route
	if method != "" {
		// Register the route with the registry
		m.registry.register(pattern, method)
		// Prepend method to pattern for mux registration
		pattern = method + " " + pattern
	}

	// Apply the middleware chain
	h := m.middleware.ThenFunc(handler)

	// Register the handler
	m.ServeMux.Handle(pattern, h)
}

func (m *Mux) handleNotFound(w http.ResponseWriter, r *http.Request) {
	if m.notFoundHandler != nil {
		m.notFoundHandler(w, r)
		return
	}
	http.NotFound(w, r)
}

func (m *Mux) handleOptions(w http.ResponseWriter, r *http.Request) {
	// Only handle OPTIONS requests, anything else is a 404
	if r.Method != http.MethodOptions {
		m.handleNotFound(w, r)
		return
	}

	methods := m.registry.getAllowedMethods(r.URL.Path)
	if len(methods) == 0 {
		m.handleNotFound(w, r)
		return
	}

	w.Header().Set("Allow", strings.Join(methods, ", "))
	w.WriteHeader(http.StatusNoContent)
}

// HandleFunc registers a handler without method restrictions
func (m *Mux) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.handle(pattern, handler)
}

// Get registers a GET handler
func (m *Mux) Get(pattern string, handler http.HandlerFunc) {
	m.handle("GET "+pattern, handler)
}

// Post registers a POST handler
func (m *Mux) Post(pattern string, handler http.HandlerFunc) {
	m.handle("POST "+pattern, handler)
}

// Put registers a PUT handler
func (m *Mux) Put(pattern string, handler http.HandlerFunc) {
	m.handle("PUT "+pattern, handler)
}

// Delete registers a DELETE handler
func (m *Mux) Delete(pattern string, handler http.HandlerFunc) {
	m.handle("DELETE "+pattern, handler)
}

// Patch registers a PATCH handler
func (m *Mux) Patch(pattern string, handler http.HandlerFunc) {
	m.handle("PATCH "+pattern, handler)
}

// Options registers an OPTIONS handler
func (m *Mux) Options(pattern string, handler http.HandlerFunc) {
	m.handle("OPTIONS "+pattern, handler)
}

// Head registers a HEAD handler
func (m *Mux) Head(pattern string, handler http.HandlerFunc) {
	m.handle("HEAD "+pattern, handler)
}

type ListInfo struct {
	Pattern string   `json:"pattern"`
	Methods []string `json:"methods"`
}

// ListRoutes returns a list of all registered routes
func (m *Mux) ListRoutes() []ListInfo {
	routes := m.registry.getRoutes()
	list := make([]ListInfo, 0, len(routes))
	for _, r := range routes {
		methods := make([]string, 0, len(r.Methods))

		for method := range r.Methods {
			methods = append(methods, method)
		}

		// Sort for consistent output
		sort.Strings(methods)

		list = append(list, ListInfo{
			Pattern: r.Pattern,
			Methods: methods,
		})
	}

	return list
}

// DumpRoutes returns a JSON representation of all routes
func (m *Mux) DumpRoutes() (string, error) {
	routes := m.ListRoutes()
	b, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}