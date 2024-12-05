package hop_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop"
	"github.com/patrickward/hop/conf"
	"github.com/patrickward/hop/route"
)

// Mock modules for testing
type mockModule struct {
	id       string
	initErr  error
	startErr error
	stopErr  error
	initDone chan struct{}
}

func (m *mockModule) ID() string { return m.id }
func (m *mockModule) Init() error {
	if m.initDone != nil {
		defer close(m.initDone)
	}
	return m.initErr
}

func (m *mockModule) Start(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("module %s: %w", m.id, ctx.Err())
	default:
		if m.startErr != nil {
			return fmt.Errorf("module %s: %w", m.id, m.startErr)
		}
		return nil
	}
}

func (m *mockModule) Stop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("module %s: %w", m.id, ctx.Err())
	default:
		if m.stopErr != nil {
			return fmt.Errorf("module %s: %w", m.id, m.stopErr)
		}
		return nil
	}
}

type mockHTTPModule struct {
	mockModule
	handlers map[string]http.HandlerFunc
}

func (m *mockHTTPModule) RegisterRoutes(router *route.Mux) {
	for pattern, handler := range m.handlers {
		router.HandleFunc(pattern, handler)
	}
}

func TestModuleRegistration(t *testing.T) {
	tests := []struct {
		name      string
		modules   []hop.Module
		wantErrs  []bool
		afterFunc func(*testing.T, *hop.App)
	}{
		{
			name: "register single module successfully",
			modules: []hop.Module{
				&mockModule{id: "test1"},
			},
			wantErrs: []bool{false},
		},
		{
			name: "register multiple modules successfully",
			modules: []hop.Module{
				&mockModule{id: "test1"},
				&mockModule{id: "test2"},
			},
			wantErrs: []bool{false, false},
		},
		{
			name: "fail when registering duplicate modules",
			modules: []hop.Module{
				&mockModule{id: "test1"},
				&mockModule{id: "test1"},
			},
			wantErrs: []bool{false, true},
		},
		{
			name: "module initialization error",
			modules: []hop.Module{
				&mockModule{id: "test1", initErr: errors.New("init failed")},
			},
			wantErrs: []bool{true},
		},
		{
			name: "verify module exists after registration",
			modules: []hop.Module{
				&mockModule{id: "test1"},
			},
			wantErrs: []bool{false},
			afterFunc: func(t *testing.T, app *hop.App) {
				module, err := app.GetModule("test1")
				assert.NoError(t, err)
				assert.NotNil(t, module)
				assert.Equal(t, "test1", module.ID())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := createTestApp(t)
			require.NoError(t, err)

			for i, module := range tt.modules {
				app.RegisterModule(module)
				err := app.Error()
				if tt.wantErrs[i] {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			if tt.afterFunc != nil {
				tt.afterFunc(t, app)
			}
		})
	}
}

func TestModuleStart(t *testing.T) {
	tests := []struct {
		name          string
		modules       []mockModule
		ctx           context.Context
		cancelContext bool
		wantErrs      []string
	}{
		{
			name: "successful start",
			modules: []mockModule{
				{id: "test1"},
				{id: "test2"},
			},
		},
		{
			name: "single module start error",
			modules: []mockModule{
				{id: "test1", startErr: errors.New("start failed")},
			},
			wantErrs: []string{"module test1: start failed"},
		},
		{
			name: "multiple module start errors",
			modules: []mockModule{
				{id: "test1", startErr: errors.New("error1")},
				{id: "test2", startErr: errors.New("error2")},
			},
			wantErrs: []string{
				"module test1: error1",
				"module test2: error2",
			},
		},
		{
			name: "context cancellation",
			modules: []mockModule{
				{id: "test1"},
			},
			cancelContext: true,
			wantErrs:      []string{"context canceled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := createTestApp(t)
			require.NoError(t, err)

			for _, m := range tt.modules {
				app.RegisterModule(&m)
				err := app.Error()
				require.NoError(t, err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if tt.cancelContext {
				cancel()
			}

			err = app.StartModules(ctx)
			if len(tt.wantErrs) > 0 {
				assert.Error(t, err)
				for _, wantErr := range tt.wantErrs {
					assert.Contains(t, err.Error(), wantErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModuleStop(t *testing.T) {
	tests := []struct {
		name          string
		modules       []mockModule
		ctx           context.Context
		cancelContext bool
		wantErrs      []string
	}{
		{
			name: "successful stop",
			modules: []mockModule{
				{id: "test1"},
				{id: "test2"},
			},
		},
		{
			name: "single module stop error",
			modules: []mockModule{
				{id: "test1", stopErr: errors.New("stop failed")},
			},
			wantErrs: []string{"module test1: stop failed"},
		},
		{
			name: "multiple module stop errors",
			modules: []mockModule{
				{id: "test1", stopErr: errors.New("error1")},
				{id: "test2", stopErr: errors.New("error2")},
			},
			wantErrs: []string{
				"module test1: error1",
				"module test2: error2",
			},
		},
		{
			name: "context cancellation",
			modules: []mockModule{
				{id: "test1"},
			},
			cancelContext: true,
			wantErrs:      []string{"context canceled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := createTestApp(t)
			require.NoError(t, err)

			for _, m := range tt.modules {
				app.RegisterModule(&m)
				err := app.Error()
				require.NoError(t, err)
			}

			// We need to start before we can stop
			err = app.StartModules(context.Background())
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if tt.cancelContext {
				cancel()
			}

			err = app.Stop(ctx)
			if len(tt.wantErrs) > 0 {
				assert.Error(t, err)
				for _, wantErr := range tt.wantErrs {
					assert.Contains(t, err.Error(), wantErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModuleLifecycle(t *testing.T) {
	tests := []struct {
		name         string
		module       mockModule
		startCtx     context.Context
		stopCtx      context.Context
		cancelStart  bool
		cancelStop   bool
		wantStartErr bool
		wantStopErr  bool
	}{
		{
			name: "successful lifecycle",
			module: mockModule{
				id:       "test1",
				initDone: make(chan struct{}),
			},
		},
		{
			name: "start error",
			module: mockModule{
				id:       "test2",
				startErr: errors.New("start failed"),
			},
			wantStartErr: true,
		},
		{
			name: "stop error",
			module: mockModule{
				id:      "test3",
				stopErr: errors.New("stop failed"),
			},
			wantStopErr: true,
		},
		{
			name: "context canceled during start",
			module: mockModule{
				id: "test4",
			},
			cancelStart:  true,
			wantStartErr: true,
		},
		{
			name: "context canceled during stop",
			module: mockModule{
				id: "test5",
			},
			cancelStop:  true,
			wantStopErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := createTestApp(t)
			require.NoError(t, err)

			app.RegisterModule(&tt.module)
			err = app.Error()
			require.NoError(t, err)

			// Start context
			startCtx, startCancel := context.WithTimeout(context.Background(), time.Second)
			defer startCancel()
			if tt.cancelStart {
				startCancel()
			}

			// Test Start
			err = app.StartModules(startCtx)
			if tt.wantStartErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Stop context
			stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
			defer stopCancel()
			if tt.cancelStop {
				stopCancel()
			}

			// Test Stop
			err = app.Stop(stopCtx)
			if tt.wantStopErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPModuleRoutes(t *testing.T) {
	tests := []struct {
		name       string
		module     hop.Module
		reqMethod  string
		reqPath    string
		wantStatus int
	}{
		{
			name: "registered route responds",
			module: &mockHTTPModule{
				mockModule: mockModule{id: "http1"},
				handlers: map[string]http.HandlerFunc{
					"/test": func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					},
				},
			},
			reqMethod:  "GET",
			reqPath:    "/test",
			wantStatus: http.StatusOK,
		},
		{
			name: "unregistered route returns 404",
			module: &mockHTTPModule{
				mockModule: mockModule{id: "http2"},
				handlers: map[string]http.HandlerFunc{
					"/test": func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					},
				},
			},
			reqMethod:  "GET",
			reqPath:    "/nonexistent",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := createTestApp(t)
			require.NoError(t, err)

			app.RegisterModule(tt.module)
			err = app.Error()
			require.NoError(t, err)

			w := newTestResponseRecorder()
			r := httptest.NewRequest(tt.reqMethod, tt.reqPath, nil)

			app.Router().ServeHTTP(w, r)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// Helper to create a test app with minimal configuration
func createTestApp(t *testing.T) (*hop.App, error) {
	t.Helper()

	cfg := hop.AppConfig{
		Config: &conf.Config{
			App: conf.AppConfig{
				Environment: "test",
			},
			Server: conf.ServerConfig{
				Port: 4444,
			},
		},
	}
	return hop.New(cfg)
}

// Custom response recorder that implements http.ResponseWriter
type testResponseRecorder struct {
	*httptest.ResponseRecorder
	closeNotify chan bool
}

func newTestResponseRecorder() *testResponseRecorder {
	return &testResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closeNotify:      make(chan bool, 1),
	}
}

func (r *testResponseRecorder) CloseNotify() <-chan bool {
	return r.closeNotify
}

type mockTemplateDataModule struct {
	mockModule
	data  map[string]any
	calls int // Track number of times OnTemplateData is called
}

func (m *mockTemplateDataModule) OnTemplateData(r *http.Request, data *map[string]any) {
	m.calls++
	for k, v := range m.data {
		(*data)[k] = v
	}
}

func TestTemplateDataModules(t *testing.T) {
	tests := []struct {
		name    string
		modules []hop.Module
		want    map[string]any // Keys we expect to find in the template data
		repeat  int            // Number of times to call NewTemplateData
	}{
		{
			name: "single module adds data",
			modules: []hop.Module{
				&mockTemplateDataModule{
					mockModule: mockModule{id: "test1"},
					data: map[string]any{
						"key1": "value1",
					},
				},
			},
			want: map[string]any{
				"key1": "value1",
			},
			repeat: 1,
		},
		{
			name: "multiple calls verify caching",
			modules: []hop.Module{
				&mockTemplateDataModule{
					mockModule: mockModule{id: "test1"},
					data: map[string]any{
						"key1": "value1",
					},
				},
			},
			want: map[string]any{
				"key1": "value1",
			},
			repeat: 3,
		},
		{
			name: "multiple modules add different data",
			modules: []hop.Module{
				&mockTemplateDataModule{
					mockModule: mockModule{id: "test1"},
					data: map[string]any{
						"key1": "value1",
					},
				},
				&mockTemplateDataModule{
					mockModule: mockModule{id: "test2"},
					data: map[string]any{
						"key2": "value2",
					},
				},
			},
			want: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			repeat: 1,
		},
		{
			name: "non-template modules are ignored",
			modules: []hop.Module{
				&mockModule{id: "regular"},
				&mockTemplateDataModule{
					mockModule: mockModule{id: "template"},
					data: map[string]any{
						"moduleData": "value1",
					},
				},
			},
			want: map[string]any{
				"moduleData": "value1",
			},
			repeat: 1,
		},
		{
			name: "nested data is properly merged",
			modules: []hop.Module{
				&mockTemplateDataModule{
					mockModule: mockModule{id: "test1"},
					data: map[string]any{
						"nested": map[string]any{
							"key1": "value1",
						},
					},
				},
				&mockTemplateDataModule{
					mockModule: mockModule{id: "test2"},
					data: map[string]any{
						"nested": map[string]any{
							"key2": "value2",
						},
					},
				},
			},
			want: map[string]any{
				"nested": map[string]any{
					"key1": "value1",
					"key2": "value2",
				},
			},
			repeat: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := createTestApp(t)
			require.NoError(t, err)

			// Register all modules and collect template data modules for verification
			var templateModules []*mockTemplateDataModule
			for _, m := range tt.modules {
				app.RegisterModule(m)
				err = app.Error()
				require.NoError(t, err)
				if tdm, ok := m.(*mockTemplateDataModule); ok {
					templateModules = append(templateModules, tdm)
				}
			}

			// Create test request
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			// Call NewTemplateData multiple times if specified
			for i := 0; i < tt.repeat; i++ {
				data := app.NewTemplateData(r)

				// Check only for our specific keys and values
				for k, want := range tt.want {
					got, exists := data[k]
					assert.True(t, exists, "key %q not found in template data", k)
					assert.Equal(t, want, got, "value mismatch for key %q", k)
				}

				// Add small delay to ensure calls are distinct
				time.Sleep(time.Millisecond)
			}

			// Verify each template module was called the expected number of times
			for _, tdm := range templateModules {
				assert.Equal(t, tt.repeat, tdm.calls,
					"template module %s called %d times, expected %d",
					tdm.ID(), tdm.calls, tt.repeat)
			}
		})
	}
}

func TestFullServerStart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping server test in short mode")
	}

	app, err := createTestApp(t)
	require.NoError(t, err)

	app.Router().HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Start(context.Background())
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test server is running
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", app.Config().Server.Port))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Or whatever status you expect

	// Shutdown server
	err = app.ShutdownServer(context.Background())
	require.NoError(t, err)

	// Check for server errors
	serverErr := <-errCh
	assert.NoError(t, serverErr)
}
