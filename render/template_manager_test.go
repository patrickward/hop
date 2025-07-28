package render_test

import (
	"bytes"
	"html/template"
	"os"
	"testing"

	"github.com/patrickward/hop/v2/render"
	"github.com/patrickward/hop/v2/render/testdata/templates"
)

func TestTemplateManagerWithDefaults(t *testing.T) {
	t.Parallel()

	// Create a new TemplateManager instance
	tm := render.NewTemplateManager(nil)

	// Check if the TemplateManager is initialized correctly
	if tm == nil {
		t.Fatal("TemplateManager should not be nil")
	}

	// Check if the default settings are applied correctly
	if tm.LayoutsDir() != "layouts" {
		t.Errorf("Expected LayoutsDir 'layouts', got '%s'", tm.LayoutsDir())
	}

	if tm.PartialsDir() != "partials" {
		t.Errorf("Expected PartialsDir 'partials', got '%s'", tm.PartialsDir())
	}

	if tm.PagesDir() != "pages" {
		t.Errorf("Expected PagesDir 'pages', got '%s'", tm.PagesDir())
	}

	if tm.ErrorsDir() != "errors" {
		t.Errorf("Expected ErrorsDir 'errors', got '%s'", tm.ErrorsDir())
	}

	if tm.BaseLayout() != "base" {
		t.Errorf("Expected BaseLayout 'base', got '%s'", tm.BaseLayout())
	}

	if tm.Extension() != ".html" {
		t.Errorf("Expected Extension '.html', got '%s'", tm.Extension())
	}

	// Check if the template functions are initialized
	if tm.FuncMap() == nil {
		t.Error("Expected FuncMap to be initialized, got nil")
	}
}

func TestTemplateManagerWithOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		layoutsDir  string
		partialsDir string
		pagesDir    string
		errorsDir   string
		baseLayout  string
		extension   string
		funcs       template.FuncMap
	}{
		{
			name:        "default settings",
			layoutsDir:  "foo/layouts",
			partialsDir: "foo/partials",
			pagesDir:    "foo/pages",
			errorsDir:   "foo/system",
			baseLayout:  "default",
			extension:   ".tmpl",
			funcs:       template.FuncMap{"customFunc": func() string { return "custom" }},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a new TemplateManager with options
			tm := render.NewTemplateManager(nil,
				render.WithLayoutsDir(tc.layoutsDir),
				render.WithPartialsDir(tc.partialsDir),
				render.WithPagesDir(tc.pagesDir),
				render.WithErrorsDir(tc.errorsDir),
				render.WithBaseLayout(tc.baseLayout),
				render.WithExtension(tc.extension),
				render.WithFuncMap(tc.funcs),
			)

			// Check if the TemplateManager is initialized correctly
			if tm == nil {
				t.Fatal("TemplateManager should not be nil")
			}

			// Additional checks can be added here to verify the settings
			if tm.LayoutsDir() != tc.layoutsDir {
				t.Errorf("Expected LayoutsDir %q, got %q", tc.layoutsDir, tm.LayoutsDir())
			}

			if tm.PartialsDir() != tc.partialsDir {
				t.Errorf("Expected PartialsDir %q, got %q", tc.partialsDir, tm.PartialsDir())
			}

			if tm.PagesDir() != tc.pagesDir {
				t.Errorf("Expected PagesDir %q, got %q", tc.pagesDir, tm.PagesDir())
			}

			if tm.ErrorsDir() != tc.errorsDir {
				t.Errorf("Expected ErrorsDir %q, got %q", tc.errorsDir, tm.ErrorsDir())
			}

			if tm.BaseLayout() != tc.baseLayout {
				t.Errorf("Expected BaseLayout %q, got %q", tc.baseLayout, tm.BaseLayout())
			}

			if tm.Extension() != tc.extension {
				t.Errorf("Expected Extension %q, got %q", tc.extension, tm.Extension())
			}

			// Check if the template functions are initialized
			if tm.FuncMap() == nil {
				t.Error("Expected FuncMap to be initialized, got nil")
			}

			if len(tm.FuncMap()) == 0 {
				t.Error("Expected FuncMap to have functions, got empty map")
			}

			if _, exists := tm.FuncMap()["customFunc"]; !exists {
				t.Error("Expected custom function 'customFunc' to be in FuncMap, but it was not found")
			}
		})
	}
}

func TestTemplateManagerParsesOrLoadsAPage(t *testing.T) {
	t.Parallel()

	// Create a new TemplateManager instance
	tm := render.NewTemplateManager(templates.FS, render.WithExtension(".gtml"))

	// Check if the templates are loaded correctly
	if tm == nil {
		t.Fatal("TemplateManager should not be nil")
	}

	// Retrieve a template page by its path
	page, err := tm.Page("pages/home")
	if err != nil {
		t.Fatalf("Failed to get page template: %v", err)
	}

	if page == nil {
		t.Fatal("Expected page template to be non-nil")
	}

	// Create a throwaway io writer to test template execution
	w := &bytes.Buffer{}

	// Execute the template with a fake data map
	err = page.ExecuteTemplate(w, "layout:base", map[string]interface{}{"Title": "Test Page"})
	if err != nil {
		t.Errorf("Failed to execute page template with data: %v", err)
	}
}

func TestTemplateManagerReloadsTemplates(t *testing.T) {
	t.Parallel()

	// Create a new TemplateManager instance
	tm := render.NewTemplateManager(templates.FS, render.WithExtension(".gtml"))

	// Check if the templates are loaded correctly
	if tm == nil {
		t.Fatal("TemplateManager should not be nil")
	}

	// Reload the templates
	err := tm.Reload()
	if err != nil {
		t.Fatalf("Failed to reload templates: %v", err)
	}

	// Verify that the templates are still accessible after reload
	page, err := tm.Page("pages/home")
	if err != nil {
		t.Fatalf("Failed to get page template after reload: %v", err)
	}

	if page == nil {
		t.Fatal("Expected page template to be non-nil after reload")
	}

	// List all templates to ensure they are still available
	templatesList := page.DefinedTemplates()
	if len(templatesList) == 0 {
		t.Fatal("Expected templates to be defined after reload, but got none")
	}

	t.Logf("Templates after reload: %v", templatesList)
}

func TestTemplateManagerCanLoadFromLocalFilesystem(t *testing.T) {
	t.Parallel()

	// Use the local filesystem for testing and find the templates under "testdata/templates"
	localFS := os.DirFS("testdata/templates")
	tm := render.NewTemplateManager(localFS, render.WithExtension(".gtml"))
	if tm == nil {
		t.Fatal("TemplateManager should not be nil")
	}

	// Load a specific page template
	page, err := tm.Page("pages/home")
	if err != nil {
		t.Fatalf("Failed to get page template: %v", err)
	}

	if page == nil {
		t.Fatal("Expected page template to be non-nil")
	}

	// Create a throwaway io writer to test template execution
	w := &bytes.Buffer{}

	// Execute the template with a fake data map
	err = page.ExecuteTemplate(w, "layout:base", map[string]interface{}{"Title": "Test Page"})
	if err != nil {
		t.Errorf("Failed to execute page template with data: %v", err)
	}
}
