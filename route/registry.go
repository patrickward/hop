package route

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
)

var emptyStruct = struct{}{}

// Route stores information about registered routes
type Route struct {
	Pattern    string              // Original pattern
	Methods    map[string]struct{} // Allowed methods
	ParamNames []string            // Names of parameters in the pattern
}

// BuildPath generates a URL path from the pattern and parameters
func (r *Route) BuildPath(params map[string]string) (string, error) {
	p := r.Pattern

	if len(r.ParamNames) != len(params) {
		return "", fmt.Errorf("incorrect number of parameters")
	}

	for _, name := range r.ParamNames {
		value, ok := params[name]
		if !ok {
			return "", fmt.Errorf("missing parameter %q", name)
		}
		p = strings.Replace(p, ":"+name, url.PathEscape(value), 1)
	}

	return p, nil
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

// cleanPattern normalizes a pattern for consistency
func cleanPattern(pattern string) string {
	if pattern == "" {
		return "/"
	}

	// Use path.Clean to normalize the path
	clean := path.Clean(pattern)

	// Ensure it starts with a slash
	if clean[0] != '/' {
		clean = "/" + clean
	}

	// Add trailing slash for consistency with ServeMux
	if len(clean) > 1 && clean[len(clean)-1] != '/' {
		clean += "/"
	}

	return clean
}

// register adds or updates a route's allowed methods
func (rr *routeRegistry) register(pattern, method string) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	cleanPath := cleanPattern(pattern)

	route, exists := rr.routes[cleanPath]
	if !exists {
		paramNames := []string{}
		parts := strings.Split(cleanPath, "/")
		for _, part := range parts {
			if strings.HasPrefix(part, ":") {
				paramNames = append(paramNames, part[1:])
			}
		}

		route = &Route{
			Pattern:    pattern,
			Methods:    make(map[string]struct{}, 4),
			ParamNames: paramNames,
		}
		rr.routes[cleanPath] = route
	}

	// Register the explicit method
	route.Methods[method] = emptyStruct

	// If registering GET, automatically support HEAD
	if method == http.MethodGet {
		route.Methods[http.MethodHead] = emptyStruct
	}

	// Invalidate the cache for this pattern
	delete(rr.methodCache, cleanPath)
}

// getAllowedMethods returns all allowed methods for a pattern
func (rr *routeRegistry) getAllowedMethods(pattern string) []string {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	cleanPath := cleanPattern(pattern)

	// Check the cache first
	if methods, ok := rr.methodCache[cleanPath]; ok {
		return methods
	}

	info, exists := rr.routes[cleanPath]
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
	rr.methodCache[cleanPath] = methods

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
