package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/patrickward/hop/templates"
)

const defaultFSKey = "DEFAULT_FS"

type Sources map[string]fs.FS

// TemplateManager is a template adapter for the HyperView framework that uses the Go html/template package.
type TemplateManager struct {
	baseLayout    string
	systemLayout  string
	extension     string
	fileSystemMap map[string]fs.FS
	logger        *slog.Logger
	funcMap       template.FuncMap
	//templates     map[string]*template.Template

	templateCache      sync.Map
	loadOnce           sync.Once
	mu                 sync.RWMutex
	layoutsAndPartials *template.Template
}

// TemplateManagerOptions are the options for the TemplateManager.
type TemplateManagerOptions struct {
	// BaseLayout is the default layout to use for rendering templates. Default is "base".
	BaseLayout string

	// SystemLayout is the layout to use for system pages (e.g. 404, 500). Default is "base".
	SystemLayout string

	// Extension is the file extension for the templates. Default is ".html".
	Extension string

	// Funcs is a map of functions to add to default set of template functions made available. See the `templates/funcmap` package for a list of default functions.
	Funcs template.FuncMap

	// Logger is the logger to use for logging errors. Default is nil.
	Logger *slog.Logger
}

// NewTemplateManager creates a new TemplateManager.
// Accepts a map of file systems, a logger, and options for configuration.
// For sources, if the string key is empty or "-", it will be treated as the default file system. Otherwise, the key is used as the file system ID.
// e.g., "foo:bar" for a template named "bar" in the "foo" file system.
func NewTemplateManager(sources Sources, opts TemplateManagerOptions) (*TemplateManager, error) {
	funcMap := templates.MergeFuncMaps(templates.FuncMap(), opts.Funcs)

	// Set default extension if not provided
	if opts.Extension == "" {
		opts.Extension = ".html"
	}

	// Ensure the extension starts with a .
	if opts.Extension[0] != '.' {
		opts.Extension = "." + opts.Extension
	}

	// If no base layout is provided, set it to "base"
	if opts.BaseLayout == "" {
		opts.BaseLayout = DefaultBaseLayout
	}

	// If no system layout is provided, set it to "base"
	if opts.SystemLayout == "" {
		opts.SystemLayout = opts.BaseLayout
	}

	// Normalize the filesystem map to use our default key
	normalizedSources := make(Sources)
	for k, v := range sources {
		if k == "" || k == "-" {
			normalizedSources[defaultFSKey] = v
		} else {
			normalizedSources[k] = v
		}
	}

	tm := &TemplateManager{
		fileSystemMap: normalizedSources,
		logger:        opts.Logger,
		baseLayout:    opts.BaseLayout,
		systemLayout:  opts.SystemLayout,
		extension:     opts.Extension,
		funcMap:       funcMap,
		templateCache: sync.Map{},
	}

	return tm, tm.Initialize()
}

// NewResponse creates a new Response instance with the TemplateManager.
func (tm *TemplateManager) NewResponse() *Response {
	return NewResponse(tm)
}

// SetErrorTemplate sets the template to use for rendering system errors.
func (tm *TemplateManager) SetErrorTemplate(layout string) {
	tm.systemLayout = layout
}

// Initialize sets up the template manager and preloads critical templates
func (tm *TemplateManager) Initialize() error {
	// Validate extension format
	if tm.extension == "" {
		tm.extension = ".html"
	}
	if tm.extension[0] != '.' {
		tm.extension = "." + tm.extension
	}

	// Load layouts and partials first - these are needed for all templates
	var err error
	tm.layoutsAndPartials, err = tm.loadLayoutsAndPartials()
	if err != nil {
		return fmt.Errorf("failed to load layouts and partials: %w", err)
	}

	// Preload critical system templates with correct extension
	systemPages := []string{"404", "500", "403", "401", "503"}
	var systemTemplates []string
	for _, page := range systemPages {
		path := tm.viewsPath(SystemDir, page) + tm.extension
		systemTemplates = append(systemTemplates, path)
	}

	for _, path := range systemTemplates {
		if _, err := tm.getTemplate(path); err != nil {
			// Log but don't fail if a system template is missing
			tm.logger.Warn("Failed to preload system template",
				slog.String("path", path),
				slog.String("error", err.Error()))
		}
	}

	return nil
}

// parseTemplatePath splits a template path into filesystem ID and relative path
func (tm *TemplateManager) parseTemplatePath(path string) (string, string) {
	parts := strings.SplitN(path, ":", 2)
	if len(parts) == 2 {
		// Handle empty prefix or "-" as default filesystem
		if parts[0] == "" || parts[0] == "-" {
			return defaultFSKey, parts[1]
		}
		return parts[0], parts[1]
	}
	return defaultFSKey, path
}

// getTemplate gets or loads a template with embedded error handling
func (tm *TemplateManager) getTemplate(path string) (*template.Template, error) {
	// Check cache first
	if tmpl, ok := tm.templateCache.Load(path); ok {
		return tmpl.(*template.Template), nil
	}

	// Find the appropriate filesystem and relative path
	fsID, relPath := tm.parseTemplatePath(path)

	fsys, ok := tm.fileSystemMap[fsID]
	if !ok {
		return nil, fmt.Errorf("%w: filesystem not found: %s", ErrTempNotFound, fsID)
	}

	// If the path doesn't end with the extension, add it
	if !strings.HasSuffix(relPath, tm.extension) {
		relPath += tm.extension
	}

	// Check if the template file exists
	if _, err := fsys.Open(relPath); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTempNotFound, relPath)
	}

	// Clone and parse the template
	tm.mu.RLock()
	tmpl, err := template.Must(tm.layoutsAndPartials.Clone()).ParseFS(fsys, relPath)
	tm.mu.RUnlock()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTempParse, err)
	}

	// Cache the template
	actual, loaded := tm.templateCache.LoadOrStore(path, tmpl)
	if loaded {
		// Another goroutine beat us to it, use their template
		return actual.(*template.Template), nil
	}

	return tmpl, nil
}

// loadLayoutsAndPartials loads the common layouts and partials from the filesystems
func (tm *TemplateManager) loadLayoutsAndPartials() (*template.Template, error) {
	commonTemplates := template.New("_common_").Funcs(tm.funcMap)

	for _, fsys := range tm.fileSystemMap {
		// First, load layouts into the common template
		layoutPath := LayoutsDir + "/*" + tm.extension
		_, err := commonTemplates.ParseFS(fsys, layoutPath)
		if err != nil {
			return nil, err
		}

		processPartials := func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && filepath.Ext(path) == tm.extension {
				fullPath := path

				// Parse the partial template in the common template
				_, err := commonTemplates.ParseFS(fsys, fullPath)
				if err != nil {
					return err
				}
			}
			return nil
		}

		// If the "partials" directory exists, parse it
		if _, err := fsys.Open(PartialsDir); err == nil {
			if err := fs.WalkDir(fsys, PartialsDir, processPartials); err != nil {
				return nil, err
			}
		}
	}

	return commonTemplates, nil
}

//func (tm *TemplateManager) LogTemplateNames() {
//	for name, tmpl := range tm.templates {
//		tm.logger.Info("Template", slog.String("name", name))
//		associatedTemplates := tmpl.Templates()
//		for _, tmpl := range associatedTemplates {
//			tm.logger.Info("    Associated", slog.String("name", tmpl.Name()))
//		}
//	}
//}

// render renders a response using the template manager
func (tm *TemplateManager) render(w http.ResponseWriter, r *http.Request, resp *Response) {
	path := resp.GetTemplatePath()
	tmpl, err := tm.getTemplate(path)
	if err != nil {
		switch {
		case errors.Is(err, ErrTempNotFound):
			tm.renderSystemError(w, r, resp, 404, err)
		case errors.Is(err, ErrTempParse):
			tm.renderSystemError(w, r, resp, 500, err)
		default:
			tm.renderSystemError(w, r, resp, 500, err)
		}
		return
	}

	buf := new(bytes.Buffer)
	layout := fmt.Sprintf("layout:%s", resp.GetTemplateLayout())
	err = tmpl.ExecuteTemplate(buf, layout, resp.PageData(r).Data())
	if err != nil {
		tm.renderSystemError(w, r, resp, 500, err)
		return
	}

	// Write response
	for key, value := range resp.GetHeaders() {
		w.Header().Set(key, value)
	}
	w.WriteHeader(resp.GetStatusCode())
	if _, err := buf.WriteTo(w); err != nil {
		tm.logger.Error("Failed to write response",
			slog.String("path", path),
			slog.String("error", err.Error()))
	}
}

// viewsPath helper function to construct template paths
func (tm *TemplateManager) viewsPath(path ...string) string {
	return fmt.Sprintf("%s/%s", ViewsDir, strings.Join(path, "/"))
}

// errorPageFromStatus returns the error page name based on the HTTP status code
func errorPageFromStatus(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "401"
	case http.StatusForbidden:
		return "403"
	case http.StatusMethodNotAllowed:
		return "405"
	case http.StatusNotFound:
		return "404"
	case http.StatusServiceUnavailable:
		return "503"
	default:
		return "500"
	}
}

// renderSystemError handles rendering of system error pages with fallback
func (tm *TemplateManager) renderSystemError(w http.ResponseWriter, r *http.Request, resp *Response, status int, originalErr error) {
	// Log the original error
	tm.logger.Error("System Error",
		slog.String("path", resp.GetTemplatePath()),
		slog.String("error", originalErr.Error()))

	// Try to render the error template
	errorPath := tm.viewsPath(SystemDir, errorPageFromStatus(status))
	errorTmpl, err := tm.getTemplate(errorPath)
	if err != nil {
		// Fallback to basic error response if error template fails
		tm.logger.Error("Failed to load error template",
			slog.String("path", errorPath),
			slog.String("error", err.Error()))
		http.Error(w, originalErr.Error(), http.StatusInternalServerError)
		return
	}

	// Render the error template
	resp.Path(errorPath).Status(status)
	buf := new(bytes.Buffer)
	layout := fmt.Sprintf("layout:%s", tm.systemLayout)
	if err := errorTmpl.ExecuteTemplate(buf, layout, resp.PageData(r).Data()); err != nil {
		// Fallback if error template rendering fails
		tm.logger.Error("Failed to render error template",
			slog.String("path", errorPath),
			slog.String("error", err.Error()))
		http.Error(w, originalErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.GetStatusCode())
	if _, err := buf.WriteTo(w); err != nil {
		tm.logger.Error("Failed to write error response",
			slog.String("path", errorPath),
			slog.String("error", err.Error()))
	}
}
