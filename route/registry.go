package route

import (
	"net/http"
	"sort"
	"sync"
)

// Route stores information about registered routes
type Route struct {
	Pattern string
	Methods map[string]bool
}

// routeRegistry tracks all registered routes and their allowed methods
type routeRegistry struct {
	mu     sync.RWMutex
	routes map[string]*Route // key is the pattern
}

func newRouteRegistry() *routeRegistry {
	return &routeRegistry{
		routes: make(map[string]*Route),
	}
}

// register adds or updates a route's allowed methods
func (rr *routeRegistry) register(pattern, method string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	info, exists := rr.routes[pattern]
	if !exists {
		info = &Route{
			Pattern: pattern,
			Methods: make(map[string]bool),
		}
		rr.routes[pattern] = info
	}

	// Register the explicit method
	info.Methods[method] = true

	// If registering GET, automatically support HEAD
	if method == http.MethodGet {
		info.Methods[http.MethodHead] = true
	}
}

// getAllowedMethods returns all allowed methods for a pattern
func (rr *routeRegistry) getAllowedMethods(pattern string) []string {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	info, exists := rr.routes[pattern]
	if !exists {
		return nil
	}

	methods := make([]string, 0, len(info.Methods))
	for method := range info.Methods {
		methods = append(methods, method)
	}

	// Sort for consistent output
	sort.Strings(methods)
	return methods
}

// getRoutes returns all registered routes
func (rr *routeRegistry) getRoutes() []Route {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	routes := make([]Route, 0, len(rr.routes))
	for _, info := range rr.routes {
		// Create a copy of the route info
		methods := make(map[string]bool)
		for k, v := range info.Methods {
			methods[k] = v
		}
		routes = append(routes, Route{
			Pattern: info.Pattern,
			Methods: methods,
		})
	}
	return routes
}
