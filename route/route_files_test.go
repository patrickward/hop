package route_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/route"
)

func TestMux_ServeDirectory(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		fsPath        string
		requests      []string
		expectedCode  int
		expectError   bool
		errorContains string
	}{
		{
			name:    "serve files from static directory",
			pattern: "/static/{file...}",
			fsPath:  "testdata",
			requests: []string{
				"/static/file1.txt",
				"/static/file2.txt",
				"/static/subdir/file3.txt",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "invalid pattern without {file...}",
			pattern:       "/static",
			fsPath:        "testdata",
			expectError:   true,
			errorContains: "must contain {file...}",
		},
		{
			name:    "request non-existent file returns 404",
			pattern: "/static/{file...}",
			fsPath:  "testdata",
			requests: []string{
				"/static/missing.txt",
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := route.New()
			dir := http.Dir(tt.fsPath)
			err := mux.ServeDirectory(tt.pattern, dir)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)

			for _, reqPath := range tt.requests {
				req := httptest.NewRequest(http.MethodGet, reqPath, nil)
				w := httptest.NewRecorder()

				mux.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCode, w.Code)

				if tt.expectedCode == http.StatusOK {
					// Get just the filename part of the request path
					//urlPrefix := strings.TrimSuffix(tt.pattern, "{file...}")
					//relPath := strings.TrimPrefix(reqPath, urlPrefix)
					//relPath = strings.TrimPrefix(relPath, "/") // Remove leading slash if present
					relPath := strings.TrimPrefix(reqPath, "/") // Remove leading slash if present

					// Read the actual file content to compare
					content, err := os.ReadFile(filepath.Join(tt.fsPath, relPath))
					require.NoError(t, err)
					assert.Equal(t, string(content), w.Body.String())
				}
			}
		})
	}
}

func TestMux_ServeDirectoryWithPrefix(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		fsPrefix      string
		fsPath        string
		requests      []string
		expectedCode  int
		expectError   bool
		errorContains string
	}{
		{
			name:     "serve files with different URL and fs paths",
			pattern:  "/files/{file...}",
			fsPrefix: "/uploads/",
			fsPath:   "testdata/uploads",
			requests: []string{
				"/files/image1.png",
				"/files/docs/image2.png",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "invalid pattern without {file...}",
			pattern:       "/files",
			fsPrefix:      "/uploads/",
			fsPath:        "testdata/uploads",
			expectError:   true,
			errorContains: "must contain {file...}",
		},
		{
			name:          "invalid fs prefix without leading slash",
			pattern:       "/files/{file...}",
			fsPrefix:      "uploads",
			fsPath:        "testdata/uploads",
			expectError:   true,
			errorContains: "must start with /",
		},
		{
			name:     "request non-existent file returns 404",
			pattern:  "/files/{file...}",
			fsPrefix: "/uploads/",
			fsPath:   "testdata/uploads",
			requests: []string{
				"/files/missing.jpg",
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := route.New()
			dir := http.Dir(tt.fsPath)
			err := mux.ServeDirectoryWithPrefix(tt.pattern, tt.fsPrefix, dir)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)

			for _, reqPath := range tt.requests {
				req := httptest.NewRequest(http.MethodGet, reqPath, nil)
				w := httptest.NewRecorder()

				mux.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCode, w.Code)

				if tt.expectedCode == http.StatusOK {
					// Read the actual file content to compare
					relPath := filepath.Clean(reqPath[len("/files/"):])
					content, err := os.ReadFile(filepath.Join(tt.fsPath, relPath))
					require.NoError(t, err)
					assert.Equal(t, string(content), w.Body.String())
				}
			}
		})
	}
}

func TestMux_ServeFiles(t *testing.T) {
	tests := []struct {
		name          string
		urlPrefix     string
		fsPath        string
		fileMappings  []any // Mix of strings and FileMapping
		requests      []string
		expectedCode  int
		expectError   bool
		errorContains string
	}{
		{
			name:      "serve individual files at root",
			urlPrefix: "",
			fsPath:    "testdata/static",
			fileMappings: []any{
				"/file1.txt",
				route.FileMapping{
					URLPath:  "/sub.txt",
					FilePath: "/subdir/file3.txt",
				},
			},
			requests: []string{
				"/file1.txt",
				"/sub.txt",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:      "serve files under URL prefix",
			urlPrefix: "/assets",
			fsPath:    "testdata/uploads",
			fileMappings: []any{
				"/image1.png",
				route.FileMapping{
					URLPath:  "/documents/doc.pdf",
					FilePath: "/docs/image2.png",
				},
			},
			requests: []string{
				"/assets/image1.png",
				"/assets/documents/doc.pdf",
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := route.New()
			dir := http.Dir(tt.fsPath)
			err := mux.ServeFiles(dir, tt.urlPrefix, tt.fileMappings...)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)

			for _, reqPath := range tt.requests {
				req := httptest.NewRequest(http.MethodGet, reqPath, nil)
				w := httptest.NewRecorder()

				mux.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCode, w.Code,
					"Expected OK status for path: %s", reqPath)

				// For this test, we're just checking status codes
				// as file content matching would require mapping each request to its source file
			}
		})
	}
}
