package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
)

// GroupFunc is a function that configures a route group
type GroupFunc func(g *Group)

// Mux extends http.ServeMux with additional routing features.
// It also provides a middleware chain for adding middleware to routes.
type Mux struct {
	*http.ServeMux
	middleware      Chain
	registry        *routeRegistry
	notFoundHandler http.Handler
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

// PrefixGroup creates a new route group with the given prefix and applies the given group configuration function.
func (m *Mux) PrefixGroup(prefix string, group GroupFunc) *Group {
	subGroup := &Group{
		mux:        m,
		prefix:     prefix,
		middleware: m.middleware,
		parent:     nil, // Root group has no parent
	}

	if group != nil {
		group(subGroup)
	}

	return subGroup
}

// Group creates a new route group with the given configuration function.
func (m *Mux) Group(group GroupFunc) *Group {
	return m.PrefixGroup("", group)
}

// Home registers a handler for the root path
func (m *Mux) Home(handler http.Handler) {
	m.handle("/{$}", handler)
}

// NotFound registers a handler for when no routes match
func (m *Mux) NotFound(handler http.Handler) {
	m.notFoundHandler = handler
}

// handle registers a handler with middleware
func (m *Mux) handle(pattern string, handler http.Handler) {
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
	h := m.middleware.Then(handler)

	// Register the handler
	m.ServeMux.Handle(pattern, h)
}

func (m *Mux) handleNotFound(w http.ResponseWriter, r *http.Request) {
	if m.notFoundHandler != nil {
		// Wrap the not found handler with the middleware chain
		h := m.middleware.Then(m.notFoundHandler)
		h.ServeHTTP(w, r)
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

// Handle registers a handler without method restrictions
func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.handle(pattern, handler)
}

// HandleFunc registers a handler without method restrictions
func (m *Mux) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.handle(pattern, handler)
}

// Get registers a GET handler
func (m *Mux) Get(pattern string, handler http.Handler) {
	m.handle("GET "+pattern, handler)
}

// Post registers a POST handler
func (m *Mux) Post(pattern string, handler http.Handler) {
	m.handle("POST "+pattern, handler)
}

// Put registers a PUT handler
func (m *Mux) Put(pattern string, handler http.Handler) {
	m.handle("PUT "+pattern, handler)
}

// Delete registers a DELETE handler
func (m *Mux) Delete(pattern string, handler http.Handler) {
	m.handle("DELETE "+pattern, handler)
}

// Patch registers a PATCH handler
func (m *Mux) Patch(pattern string, handler http.Handler) {
	m.handle("PATCH "+pattern, handler)
}

// Options registers an OPTIONS handler
func (m *Mux) Options(pattern string, handler http.Handler) {
	m.handle("OPTIONS "+pattern, handler)
}

// Head registers a HEAD handler
func (m *Mux) Head(pattern string, handler http.Handler) {
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

// -----------------------------------------------------------------------------
// Static file serving
// -----------------------------------------------------------------------------

// ServeDirectory serves all files under a filesystem directory, matching the URL paths directly.
// The provided pattern must use Go 1.22's enhanced patterns, e.g. "/static/{file...}"
//
// Requirements:
//
//   - pattern must contain the wildcard pattern {file...} (e.g. "/static/{file...}")
//   - fs cannot be nil
//
// Returns an error if the pattern is invalid or missing the {file...} suffix.
func (m *Mux) ServeDirectory(pattern string, fs http.FileSystem) error {
	if fs == nil {
		return fmt.Errorf("filesystem cannot be nil")
	}

	if !strings.Contains(pattern, "{file...}") {
		return fmt.Errorf("pattern must contain {file...} to match file paths")
	}

	fileServer := http.FileServer(fs)
	m.ServeMux.Handle(pattern, fileServer)
	return nil
}

// ServeDirectoryWithPrefix serves files that exist under fsPrefix in the filesystem
// at URLs matching the provided pattern. It requires Go 1.22's enhanced patterns to indicate file paths.
//
// Requirements:
//   - pattern must contain the wildcard pattern {file...} (e.g. "/foo/{file...}")
//   - fsPrefix must start with "/" and not end with "/"
//   - fs cannot be nil
//
// Example: ServeDirectoryWithPrefix("/foo/{file...}", "/uploads", fs)
// will serve files that exist at "/uploads/image.jpg" in the filesystem
// when requested at "/foo/image.jpg"
//
// Returns an error if any of the requirements are not met.
func (m *Mux) ServeDirectoryWithPrefix(pattern string, fsPrefix string, fs http.FileSystem) error {
	// Validate inputs
	if fs == nil {
		return fmt.Errorf("filesystem cannot be nil")
	}

	if !strings.Contains(pattern, "{file...}") {
		return fmt.Errorf("pattern must contain {file...} to match file paths")
	}

	if !strings.HasPrefix(fsPrefix, "/") {
		return fmt.Errorf("fsPrefix must start with /")
	}

	// Get the URL prefix from the pattern (everything before {file...})
	patternParts := strings.Split(pattern, "{file...}")
	if len(patternParts) != 2 {
		return fmt.Errorf("invalid pattern format")
	}
	urlPrefix := patternParts[0]

	// Create the FileServer with prefix stripping
	fileServer := http.StripPrefix(fsPrefix, http.FileServer(fs))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip the URL prefix to get the file path
		path := strings.TrimPrefix(r.URL.Path, urlPrefix)

		// Create new request with path under fsPrefix
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = filepath.Join(fsPrefix, path)

		// Serve the file
		fileServer.ServeHTTP(w, r2)
	})

	m.ServeMux.Handle(pattern, handler)
	return nil
}

// FileMapping represents a mapping between a URL path and a filesystem path
type FileMapping struct {
	URLPath  string // The URL path where the file will be served
	FilePath string // The path to the file within the filesystem
}

// ServeFiles registers multiple individual files from a filesystem at their respective paths.
// Files can be specified using their full path within the filesystem.
// If urlPrefix is provided (e.g., "/assets"), files will be served under that URL path.
// If urlPrefix is empty (""), files will be served at the root level.
// The paths argument can be a mix of strings and FileMapping structs.
//
// Examples:
//
// 1. Serve files at root level
// router.ServeFiles(http.FS(root.Files), "",  // empty prefix serves at root
//
//	"/foo/bar.png",      // Serves at /bar.png
//	"/icons/main.png",   // Serves at /main.png
//
// )
//
// 2. Serve files under a URL prefix
// router.ServeFiles(http.FS(root.Files), "/assets",
//
//	"/foo/bar.png",      // Serves at /assets/bar.png
//	"/icons/main.png",   // Serves at /assets/main.png
//
// )
//
// 3. Mix and match with custom URL paths
// router.ServeFiles(http.FS(root.Files), "/special",
//
//	"/foo/bar.png",      // Serves at /special/bar.png
//	FileMapping{         // Custom URL path still respects prefix
//		URLPath: "icons/custom.png",
//		FilePath: "/icons/main.png",
//	},                   // Serves at /special/icons/custom.png
//
// )
//
// 4. Use mappings at root level
// router.ServeFiles(http.FS(root.Files), "",
//
//	FileMapping{
//		URLPath: "/site-icon.png",
//		FilePath: "/icons/main.png",
//	},
//
// )
func (m *Mux) ServeFiles(fs http.FileSystem, urlPrefix string, paths ...any) error {
	for _, p := range paths {
		var urlPath, filePath string

		switch v := p.(type) {
		case string:
			// If just a string is provided, construct URL path from prefix and filename
			filePath = v
			fileName := filepath.Base(v)
			if urlPrefix == "" {
				urlPath = "/" + fileName
			} else {
				urlPath = filepath.Join(urlPrefix, fileName)
			}
		case FileMapping:
			// If a FileMapping is provided, respect the specified URL path but still apply prefix
			filePath = v.FilePath
			if urlPrefix == "" {
				urlPath = v.URLPath
			} else {
				urlPath = filepath.Join(urlPrefix, v.URLPath)
			}
		default:
			return fmt.Errorf("invalid path type: %T", p)
		}

		// Create a closure to capture the file path
		handler := func(fPath string) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				f, err := fs.Open(fPath)
				if err != nil {
					http.NotFound(w, r)
					return
				}
				defer func(f http.File) {
					_ = f.Close()
				}(f)

				stat, err := f.Stat()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.ServeContent(w, r, filepath.Base(fPath), stat.ModTime(), f)
			}
		}(filePath)

		// Register the handler directly with ServeMux
		m.ServeMux.HandleFunc(urlPath, handler)
	}

	return nil
}

// ServeFileFrom serves a single file from a filesystem at a specific URL path.
// This is useful when you want to serve a specific file at a custom URL path.
//
// Example:
//
// mux.ServeFileFrom("/favicon.ico", http.Dir("static"), "favicon.ico")
// This will serve the "favicon.ico" file from the "static" directory at the "/favicon.ico" URL path.
func (m *Mux) ServeFileFrom(urlPath string, fs http.FileSystem, filePath string) {
	m.ServeMux.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
		f, err := fs.Open(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer func(f http.File) {
			_ = f.Close()
		}(f)

		stat, err := f.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, filePath, stat.ModTime(), f)
	})
}

// Path generates a URL path for a route pattern without parameters.
func (m *Mux) Path(pattern string) (string, error) {
	route, exists := m.registry.routes[cleanPattern(pattern)]
	if !exists {
		return "", fmt.Errorf("route pattern %q not found", pattern)
	}

	if len(route.ParamNames) > 0 {
		return "", fmt.Errorf("route pattern %q requires parameters - use PathWithParams instead", pattern)
	}

	return route.Pattern, nil
}

// MustPath is like Path but panics if the route doesn't exist.
// It should only be used for routes without parameters.
func (m *Mux) MustPath(pattern string) string {
	path, err := m.Path(pattern)
	if err != nil {
		panic(fmt.Sprintf("failed to build path: %v", err))
	}
	return path
}

// PathWithParams generates a URL path for a route pattern with parameters.
func (m *Mux) PathWithParams(pattern string, params map[string]string) (string, error) {
	route, exists := m.registry.routes[cleanPattern(pattern)]
	if !exists {
		return "", fmt.Errorf("route pattern %q not found", pattern)
	}
	return route.BuildPath(params)
}

// MustPathWithParams is like PathWithParams but panics if the route doesn't exist
// or if required parameters are missing.
func (m *Mux) MustPathWithParams(pattern string, params map[string]string) string {
	path, err := m.PathWithParams(pattern, params)
	if err != nil {
		panic(fmt.Sprintf("failed to build path: %v", err))
	}
	return path
}

// VerifyRoute checks if a route pattern exists and supports the given method
func (m *Mux) VerifyRoute(pattern, method string) bool {
	route, exists := m.registry.routes[cleanPattern(pattern)]
	if !exists {
		return false
	}
	_, methodAllowed := route.Methods[method]
	return methodAllowed
}
