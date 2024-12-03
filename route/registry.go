package route

import (
	"net/http"
	"sort"
	"sync"
)

var emptyStruct = struct{}{}

// Route stores information about registered routes
type Route struct {
	Pattern string
	Methods map[string]struct{}
}

// routeRegistry tracks all registered routes and their allowed methods
type routeRegistry struct {
	mu          sync.RWMutex
	routes      map[string]*Route   // Key is the pattern
	methodCache map[string][]string // Cache common HTTP method too avoid allocations
}

func newRouteRegistry() *routeRegistry {
	return &routeRegistry{
		routes:      make(map[string]*Route),
		methodCache: make(map[string][]string),
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
			Methods: make(map[string]struct{}, 4),
		}
		rr.routes[pattern] = info
	}

	// Register the explicit method
	info.Methods[method] = emptyStruct

	// If registering GET, automatically support HEAD
	if method == http.MethodGet {
		info.Methods[http.MethodHead] = emptyStruct
	}

	// Invalidate the cache for this pattern
	delete(rr.methodCache, pattern)
}

// getAllowedMethods returns all allowed methods for a pattern
func (rr *routeRegistry) getAllowedMethods(pattern string) []string {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	// Check the cache first
	if methods, ok := rr.methodCache[pattern]; ok {
		return methods
	}

	info, exists := rr.routes[pattern]
	if !exists {
		return nil
	}

	// Create new slice with capacity matching methods
	methods := make([]string, 0, len(info.Methods))
	for method := range info.Methods {
		methods = append(methods, method)
	}

	// Sort for consistent output
	sort.Strings(methods)

	// Update the cache
	rr.methodCache[pattern] = methods

	return methods
}

// getRoutes returns all registered routes
func (rr *routeRegistry) getRoutes() []Route {
	rr.mu.RLock()
	defer rr.mu.RUnlock()

	routes := make([]Route, 0, len(rr.routes))
	for _, info := range rr.routes {
		// Create a copy of the route info
		methods := make(map[string]struct{})
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
