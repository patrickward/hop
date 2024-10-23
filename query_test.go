package httpgo_test

import (
	"net/http/httptest"
	"testing"

	"github.com/patrickward/httpgo"
)

func TestQueryString(t *testing.T) {
	tests := []struct {
		name     string
		queryVal string
		expVal   string
	}{
		{"Empty String", "", ""},
		{"Normal String", "GoLand", "GoLand"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?key="+tt.queryVal, nil)
			val := httpgo.QueryString(req, "key")
			if val != tt.expVal {
				t.Errorf("expected %s, got %s", tt.expVal, val)
			}
		})
	}
}

func TestQueryInt64(t *testing.T) {
	tests := []struct {
		name     string
		queryVal string
		expVal   int64
	}{
		{"Empty String", "", 0},
		{"Invalid Int64", "GoLand", 0},
		{"Valid Int64", "123456789", 123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?key="+tt.queryVal, nil)
			val := httpgo.QueryInt64(req, "key")
			if val != tt.expVal {
				t.Errorf("expected %d, got %d", tt.expVal, val)
			}
		})
	}
}

func TestQueryFloat64(t *testing.T) {
	tests := []struct {
		name     string
		queryVal string
		expVal   float64
	}{
		{"Empty String", "", 0},
		{"Invalid Float", "GoLand", 0},
		{"Valid Float", "123.456", 123.456},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?key="+tt.queryVal, nil)
			val := httpgo.QueryFloat64(req, "key")
			if val != tt.expVal {
				t.Errorf("expected %f, got %f", tt.expVal, val)
			}
		})
	}
}

func TestQueryBool(t *testing.T) {
	tests := []struct {
		name     string
		queryVal string
		expVal   bool
	}{
		{"Empty String", "", false},
		{"Invalid Bool", "GoLand", false},
		{"Valid Bool (True)", "true", true},
		{"Valid Bool (False)", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?key="+tt.queryVal, nil)
			val := httpgo.QueryBool(req, "key")
			if val != tt.expVal {
				t.Errorf("expected %t, got %t", tt.expVal, val)
			}
		})
	}
}
