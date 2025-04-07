package view

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	"github.com/patrickward/hop/v2/flash"
)

// TemplateManager is a template adapter for the HyperView framework that uses the Go html/template package.
type TemplateManager struct {
	baseLayout    string             // The default layout to use for rendering templates
	systemLayout  string             // The layout to use for system pages (e.g. 404, 500)
	extension     string             // The file extension for templates
	fs            fs.FS              // The filesystem to use for templates
	logger        *slog.Logger       // The logger to use for errors
	funcMap       template.FuncMap   // The functions to make available to templates
	templateCache sync.Map           // Cache of parsed templates
	baseTemplate  *template.Template // The base template with layouts and partials
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

// MergeIntoFuncMap merges the provided function maps into the provided function map.
func MergeIntoFuncMap(dst template.FuncMap, maps ...template.FuncMap) {
	for _, src := range maps {
		for key, value := range src {
			dst[key] = value
		}
	}
}

// NewTemplateManager creates a new TemplateManager.
// Accepts a file systems, a logger, and options for configuration.
func NewTemplateManager(fs fs.FS, opts TemplateManagerOptions) (*TemplateManager, error) {
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

	tm := &TemplateManager{
		fs:            fs,
		logger:        opts.Logger,
		baseLayout:    opts.BaseLayout,
		systemLayout:  opts.SystemLayout,
		extension:     opts.Extension,
		funcMap:       opts.Funcs,
		templateCache: sync.Map{},
	}

	return tm, tm.Initialize()
}

// NewResponse creates a new Response instance with the TemplateManager.
func (tm *TemplateManager) NewResponse(manager *flash.Manager) *Response {
	return NewResponse(tm).SetFlashManager(manager)
}

// SetSystemPagesLayout sets the layout template to use for rendering system/error pages.
func (tm *TemplateManager) SetSystemPagesLayout(layout string) {
	tm.systemLayout = layout
}

// Initialize sets up the template manager and preloads critical templates
func (tm *TemplateManager) Initialize() error {
	base := template.New("").Funcs(tm.funcMap)

	// Load all layouts
	layoutFiles, err := fs.Glob(tm.fs, filepath.Join(LayoutsDir, "*"+tm.extension))
	if err != nil {
		return fmt.Errorf("failed to load layouts: %w", err)
	}

	if len(layoutFiles) > 0 {
		_, err = base.ParseFS(tm.fs, layoutFiles...)
		if err != nil {
			return fmt.Errorf("failed to parse layouts: %w", err)
		}
	}

	// Load all partials
	err = fs.WalkDir(tm.fs, PartialsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == tm.extension {
			if _, err := base.ParseFS(tm.fs, path); err != nil {
				return fmt.Errorf("failed to parse partial: %w", err)
			}
		}

		return nil
	})

	// Only return error if it's not a "not exist" error
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrTempParse, err)
	}

	tm.baseTemplate = base
	return nil
}

// GetBaseTemplate returns the base template, which contains layouts and partials
func (tm *TemplateManager) GetBaseTemplate() (*template.Template, error) {
	if tm.baseTemplate == nil {
		return nil, fmt.Errorf("base template not initialized")
	}
	return tm.baseTemplate, nil
}

// getTemplate gets or loads a template with embedded error handling
func (tm *TemplateManager) getTemplate(path string) (*template.Template, error) {
	// Check cache first
	if cached, ok := tm.templateCache.Load(path); ok {
		return cached.(*template.Template), nil
	}

	// Ensure path has extension
	if !strings.HasSuffix(path, tm.extension) {
		path += tm.extension
	}

	// Check if template exists
	if _, err := tm.fs.Open(path); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTempNotFound, path)
	}

	// Clone base template and parse the requested template
	tmpl, err := tm.baseTemplate.Clone()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTempParse, err)
	}

	if _, err := tmpl.ParseFS(tm.fs, path); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTempParse, err)
	}

	// Cache the result
	actual, loaded := tm.templateCache.LoadOrStore(path, tmpl)
	if loaded {
		// Another goroutine beat us to it, use their template
		return actual.(*template.Template), nil
	}

	return tmpl, nil
}
