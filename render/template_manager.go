package render

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

type TemplateManager struct {
	layoutsDir    string             // Directory for layout templates
	partialsDir   string             // Directory for partial templates
	pagesDir      string             // Directory for view templates
	errorsDir     string             // Directory for error templates within the pages directory
	baseLayout    string             // The default layout to use for rendering templates
	errorsLayout  string             // The default layout to use for error templates
	extension     string             // The file extension for templates
	fs            fs.FS              // The filesystem to use for templates
	funcMap       template.FuncMap   // The functions to make available to templates
	templateCache sync.Map           // Cache of parsed templates
	baseTemplate  *template.Template // The base template with layouts and partials
	reloadMu      sync.Mutex         // Mutex to protect concurrent reloads
}

// TypeManagerOption is a function that configures the TemplateManager using the functional options pattern.
type TypeManagerOption func(*TemplateManager)

// NewTemplateManager creates a new TemplateManager with default settings.
func NewTemplateManager(fs fs.FS, opts ...TypeManagerOption) *TemplateManager {
	tm := &TemplateManager{
		layoutsDir:    "layouts",
		partialsDir:   "partials",
		pagesDir:      "pages",
		errorsDir:     "errors",
		baseLayout:    "base",
		extension:     ".html",
		fs:            fs,
		funcMap:       template.FuncMap{},
		templateCache: sync.Map{},
	}

	// Set the default errors layout to match the base layout
	tm.errorsLayout = tm.baseLayout

	// Apply options to the TemplateManager
	for _, opt := range opts {
		opt(tm)
	}

	// Ensure the extension starts with a dot
	if tm.extension[0] != '.' {
		tm.extension = "." + tm.extension
	}

	return tm
}

// WithLayoutsDir sets the directory for layout templates to something other than the default "layouts" directory.
func WithLayoutsDir(dir string) TypeManagerOption {
	return func(tm *TemplateManager) {
		if dir != "" {
			tm.layoutsDir = dir
		}
	}
}

// WithPartialsDir sets the directory for partial templates to something other than the default "partials" directory.
func WithPartialsDir(dir string) TypeManagerOption {
	return func(tm *TemplateManager) {
		if dir != "" {
			tm.partialsDir = dir
		}
	}
}

// WithPagesDir sets the directory for view templates to something other than the default "pages" directory.
func WithPagesDir(dir string) TypeManagerOption {
	return func(tm *TemplateManager) {
		if dir != "" {
			tm.pagesDir = dir
		}
	}
}

// WithErrorsDir sets the directory for error templates to something other than the default "errors" directory.
// Note this directory is considered relative to the pages' directory.
func WithErrorsDir(dir string) TypeManagerOption {
	return func(tm *TemplateManager) {
		if dir != "" {
			tm.errorsDir = dir
		}
	}
}

// WithBaseLayout sets the default layout to use for rendering templates.
func WithBaseLayout(layout string) TypeManagerOption {
	return func(tm *TemplateManager) {
		if layout != "" {
			tm.baseLayout = layout
		}
	}
}

// WithErrorsLayout sets the default layout to use for error templates.
func WithErrorsLayout(layout string) TypeManagerOption {
	return func(tm *TemplateManager) {
		if layout != "" {
			tm.errorsLayout = layout
		}
	}
}

// WithExtension sets the file extension for templates to something other than the default ".html".
func WithExtension(ext string) TypeManagerOption {
	return func(tm *TemplateManager) {
		if ext != "" {
			if ext[0] != '.' {
				ext = "." + ext
			}
			tm.extension = ext
		}
	}
}

// WithFuncMap adds functions to the template function map.
func WithFuncMap(funcs template.FuncMap) TypeManagerOption {
	return func(tm *TemplateManager) {
		if tm.funcMap == nil {
			tm.funcMap = make(template.FuncMap)
		}
		for key, value := range funcs {
			tm.funcMap[key] = value
		}
	}
}

// LayoutsDir returns the directory for layout templates.
func (tm *TemplateManager) LayoutsDir() string {
	return tm.layoutsDir
}

// PartialsDir returns the directory for partial templates.
func (tm *TemplateManager) PartialsDir() string {
	return tm.partialsDir
}

// PagesDir returns the directory for view templates.
func (tm *TemplateManager) PagesDir() string {
	return tm.pagesDir
}

// ErrorsDir returns the directory for error templates.
func (tm *TemplateManager) ErrorsDir() string {
	return tm.errorsDir
}

// BaseLayout returns the default layout to use for rendering templates.
func (tm *TemplateManager) BaseLayout() string {
	return tm.baseLayout
}

// Extension returns the file extension for templates.
func (tm *TemplateManager) Extension() string {
	return tm.extension
}

// FuncMap returns the function map for templates.
func (tm *TemplateManager) FuncMap() template.FuncMap {
	if tm.funcMap == nil {
		tm.funcMap = make(template.FuncMap)
	}
	return tm.funcMap
}

// Load loads the base template, which includes all layouts and partials.
func (tm *TemplateManager) Load() error {
	base := template.New("").Funcs(tm.funcMap)

	// Load all layouts
	layoutFiles, err := fs.Glob(tm.fs, filepath.Join(tm.layoutsDir, "*"+tm.extension))
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
	err = fs.WalkDir(tm.fs, tm.partialsDir, func(path string, d fs.DirEntry, err error) error {
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

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrTemplateNotParsed, err)
	}

	tm.baseTemplate = base
	return nil
}

// Reload reloads the base template and clears the template cache.
func (tm *TemplateManager) Reload() error {
	tm.reloadMu.Lock()
	defer tm.reloadMu.Unlock()

	// Clear the template cache
	tm.templateCache = sync.Map{}

	// Load the base template again
	return tm.Load()
}

// Page retrieves a template for the given path, ensuring it is parsed and cached.
func (tm *TemplateManager) Page(path string) (*template.Template, error) {
	// Ensure the base template is loaded
	if tm.baseTemplate == nil {
		if err := tm.Load(); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrTemplateNotParsed, err)
		}
	}

	// Ensure path has extension
	if !strings.HasSuffix(path, tm.extension) {
		path += tm.extension
	}

	// Check cache first
	if cached, ok := tm.templateCache.Load(path); ok {
		return cached.(*template.Template), nil
	}

	// Check if template exists
	if _, err := tm.fs.Open(path); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, path)
	}

	// Clone the base template and parse the requested template
	tmpl, err := tm.baseTemplate.Clone()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotParsed, err)
	}

	if _, err := tmpl.ParseFS(tm.fs, path); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotParsed, err)
	}

	tmpl.Name()

	// Cache the result
	actual, loaded := tm.templateCache.LoadOrStore(path, tmpl)
	if loaded {
		// Another goroutine beat us to it, use their template
		return actual.(*template.Template), nil
	}

	return tmpl, nil
}
