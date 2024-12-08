package conf_test

import (
	"os"
	"testing"

	"github.com/patrickward/hop/conf"
)

//	func TestEnvParser(t *testing.T) {
//		tests := []struct {
//			name    string
//			env     map[string]string
//			config  interface{}
//			check   func(t *testing.T, config interface{})
//			wantErr bool
//		}{
//			{
//				name: "basic fields",
//				env: map[string]string{
//					"SERVER_HOST": "localhost",
//					"SERVER_PORT": "8080",
//				},
//				config: &struct {
//					Server struct {
//						Host string `json:"host"`
//						Port int    `json:"port"`
//					}
//				}{},
//				check: func(t *testing.T, config interface{}) {
//					cfg := config.(*struct {
//						Server struct {
//							Host string `json:"host"`
//							Port int    `json:"port"`
//						}
//					})
//					if cfg.Server.Host != "localhost" {
//						t.Errorf("got host %s, want localhost", cfg.Server.Host)
//					}
//					if cfg.Server.Port != 8080 {
//						t.Errorf("got port %d, want 8080", cfg.Server.Port)
//					}
//				},
//			},
//			{
//				name: "nested structs",
//				env: map[string]string{
//					"DATABASE_CONNECTIONS_MAX":  "100",
//					"DATABASE_CONNECTIONS_IDLE": "10",
//				},
//				config: &struct {
//					Database struct {
//						Connections struct {
//							Max  int `json:"max"`
//							Idle int `json:"idle"`
//						}
//					}
//				}{},
//				check: func(t *testing.T, config interface{}) {
//					cfg := config.(*struct {
//						Database struct {
//							Connections struct {
//								Max  int `json:"max"`
//								Idle int `json:"idle"`
//							}
//						}
//					})
//					if cfg.Database.Connections.Max != 100 {
//						t.Errorf("got max connections %d, want 100", cfg.Database.Connections.Max)
//					}
//					if cfg.Database.Connections.Idle != 10 {
//						t.Errorf("got idle connections %d, want 10", cfg.Database.Connections.Idle)
//					}
//				},
//			},
//			{
//				name: "duration parsing",
//				env: map[string]string{
//					"TIMEOUT": "30s",
//				},
//				config: &struct {
//					Timeout conf.Duration `json:"timeout"`
//				}{},
//				check: func(t *testing.T, config interface{}) {
//					cfg := config.(*struct {
//						Timeout conf.Duration `json:"timeout"`
//					})
//					want := conf.Duration{Duration: time.Second * 30}
//					if cfg.Timeout != want {
//						t.Errorf("got timeout %v, want %v", cfg.Timeout, want)
//					}
//				},
//			},
//			{
//				name: "camelCase conversion",
//				env: map[string]string{
//					"MAX_CONN_TIME":        "5",
//					"MIN_IDLE_CONNECTIONS": "2",
//				},
//				config: &struct {
//					MaxConnTime        int `json:"maxConnTime"`
//					MinIdleConnections int `json:"minIdleConnections"`
//				}{},
//				check: func(t *testing.T, config interface{}) {
//					cfg := config.(*struct {
//						MaxConnTime        int `json:"maxConnTime"`
//						MinIdleConnections int `json:"minIdleConnections"`
//					})
//					if cfg.MaxConnTime != 5 {
//						t.Errorf("got maxConnTime %d, want 5", cfg.MaxConnTime)
//					}
//					if cfg.MinIdleConnections != 2 {
//						t.Errorf("got minIdleConnections %d, want 2", cfg.MinIdleConnections)
//					}
//				},
//			},
//		}
//
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				// Set up environment
//				for k, v := range tt.env {
//					os.Setenv(k, v)
//				}
//				defer func() {
//					// Clean up environment
//					for k := range tt.env {
//						os.Unsetenv(k)
//					}
//				}()
//
//				// Parse environment variables
//				parser := conf.NewEnvParser("")
//				err := parser.Parse(tt.config)
//
//				// Check error
//				if (err != nil) != tt.wantErr {
//					t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
//					return
//				}
//
//				// Run checks
//				if tt.check != nil {
//					tt.check(t, tt.config)
//				}
//			})
//		}
//	}
//
//	func TestWithEnvPrefix(t *testing.T) {
//		tests := []struct {
//			name    string
//			prefix  string
//			env     map[string]string
//			config  interface{}
//			check   func(t *testing.T, config interface{})
//			wantErr bool
//		}{
//			{
//				name:   "basic fields",
//				prefix: "APP",
//				env: map[string]string{
//					"APP_SERVER_HOST": "localhost",
//					"APP_SERVER_PORT": "8080",
//				},
//				config: &struct {
//					Server struct {
//						Host string `json:"host"`
//						Port int    `json:"port"`
//					}
//				}{},
//				check: func(t *testing.T, config interface{}) {
//					cfg := config.(*struct {
//						Server struct {
//							Host string `json:"host"`
//							Port int    `json:"port"`
//						}
//					})
//					if cfg.Server.Host != "localhost" {
//						t.Errorf("got host %s, want localhost", cfg.Server.Host)
//					}
//					if cfg.Server.Port != 8080 {
//						t.Errorf("got port %d, want 8080", cfg.Server.Port)
//					}
//				},
//			},
//			{
//				name:   "nested structs",
//				prefix: "APP",
//				env: map[string]string{
//					"APP_DATABASE_CONNECTIONS_MAX":  "100",
//					"APP_DATABASE_CONNECTIONS_IDLE": "10",
//				},
//				config: &struct {
//					Database struct {
//						Connections struct {
//							Max  int `json:"max"`
//							Idle int `json:"idle"`
//						}
//					}
//				}{},
//				check: func(t *testing.T, config interface{}) {
//					cfg := config.(*struct {
//						Database struct {
//							Connections struct {
//								Max  int `json:"max"`
//								Idle int `json:"idle"`
//							}
//						}
//					})
//					if cfg.Database.Connections.Max != 100 {
//						t.Errorf("got max connections %d, want 100", cfg.Database.Connections.Max)
//					}
//					if cfg.Database.Connections.Idle != 10 {
//						t.Errorf("got idle connections %d, want 10", cfg.Database.Connections.Idle)
//					}
//				},
//			},
//			{
//				name:   "duration parsing",
//				prefix: "APP",
//				env: map[string]string{
//					"APP_TIMEOUT": "30s",
//				},
//				config: &struct {
//					Timeout conf.Duration `json:"timeout"`
//				}{},
//				check: func(t *testing.T, config interface{}) {
//					cfg := config.(*struct {
//						Timeout conf.Duration `json:"timeout"`
//					})
//					want := conf.Duration{Duration: time.Second * 30}
//					if cfg.Timeout != want {
//						t.Errorf("got timeout %v, want %v", cfg.Timeout, want)
//					}
//				},
//			},
//		}
//
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				// Set up environment
//				for k, v := range tt.env {
//					_ = os.Setenv(k, v)
//				}
//				defer func() {
//					// Clean up environment
//					for k := range tt.env {
//						_ = os.Unsetenv(k)
//					}
//				}()
//
//				// Parse environment variables
//				parser := conf.NewEnvParser(tt.prefix)
//				err := parser.Parse(tt.config)
//
//				// Check error
//				if (err != nil) != tt.wantErr {
//					t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
//					return
//				}
//
//				// Run checks
//				if tt.check != nil {
//					tt.check(t, tt.config)
//				}
//			})
//		}
//	}

func TestEnvParser_Parse(t *testing.T) {
	type Config struct {
		Server struct {
			Host string
			Port int
		}
		Database struct {
			URI      string
			MaxConns int
		}
	}

	tests := []struct {
		name      string
		namespace string
		env       map[string]string
		want      Config
	}{
		{
			name:      "without prefix",
			namespace: "",
			env: map[string]string{
				"SERVER_HOST":        "localhost",
				"SERVER_PORT":        "8080",
				"DATABASE_URI":       "postgres://localhost",
				"DATABASE_MAX_CONNS": "10", // Fixed: MAX_CONNS instead of MAX_CONN
			},
			want: Config{
				Server: struct {
					Host string
					Port int
				}{
					Host: "localhost",
					Port: 8080,
				},
				Database: struct {
					URI      string
					MaxConns int
				}{
					URI:      "postgres://localhost",
					MaxConns: 10,
				},
			},
		},
		{
			name:      "with prefix",
			namespace: "APP",
			env: map[string]string{
				"APP_SERVER_HOST":        "localhost",
				"APP_SERVER_PORT":        "8080",
				"APP_DATABASE_URI":       "postgres://localhost",
				"APP_DATABASE_MAX_CONNS": "10", // Fixed: MAX_CONNS instead of MAX_CONN
			},
			want: Config{
				Server: struct {
					Host string
					Port int
				}{
					Host: "localhost",
					Port: 8080,
				},
				Database: struct {
					URI      string
					MaxConns int
				}{
					URI:      "postgres://localhost",
					MaxConns: 10,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment before test
			os.Clearenv()

			// Set up environment variables
			for k, v := range tt.env {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("failed to set env var %s: %v", k, err)
				}
			}

			// Create config and parser
			cfg := &Config{}
			parser := conf.NewEnvParser(tt.namespace)

			// Parse environment
			if err := parser.Parse(cfg); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Check results
			if cfg.Server.Host != tt.want.Server.Host {
				t.Errorf("Server.Host = %v, want %v", cfg.Server.Host, tt.want.Server.Host)
			}
			if cfg.Server.Port != tt.want.Server.Port {
				t.Errorf("Server.Port = %v, want %v", cfg.Server.Port, tt.want.Server.Port)
			}
			if cfg.Database.URI != tt.want.Database.URI {
				t.Errorf("Database.URI = %v, want %v", cfg.Database.URI, tt.want.Database.URI)
			}
			if cfg.Database.MaxConns != tt.want.Database.MaxConns {
				t.Errorf("Database.MaxConns = %v, want %v", cfg.Database.MaxConns, tt.want.Database.MaxConns)
			}
		})
	}
}

// Add debug test to see exactly what environment variables we're looking for
func TestEnvNameGeneration(t *testing.T) {
	type Config struct {
		Database struct {
			MaxConns int
		}
	}

	cfg := &Config{}
	parser := conf.NewEnvParser("APP")

	// Add a debug hook to see what environment names are being generated
	t.Logf("Field name 'MaxConns' converts to: %s", conf.ToScreamingSnake("MaxConns"))

	if err := parser.Parse(cfg); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}
func TestToScreamingSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MaxConnTime", "MAX_CONN_TIME"},
		{"DatabaseURL", "DATABASE_URL"},
		{"APIKey", "API_KEY"},
		{"Simple", "SIMPLE"},
		{"HTTPServer", "HTTP_SERVER"},
		{"minConnections", "MIN_CONNECTIONS"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := conf.ToScreamingSnake(tt.input)
			if got != tt.want {
				t.Errorf("toScreamingSnake(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
