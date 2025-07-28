package htmx_test

import (
	"net/http"
	"testing"

	"github.com/patrickward/hop/v2/render/htmx"
)

func TestIsHtmxRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{"no headers", map[string]string{}, false},
		{"only HXRequest true", map[string]string{htmx.HXRequest: "true"}, true},
		{"only HXBoosted true", map[string]string{htmx.HXBoosted: "true"}, false},
		{"both HXRequest and HXBoosted true", map[string]string{htmx.HXRequest: "true", htmx.HXBoosted: "true"}, false},
		{"invalid values", map[string]string{htmx.HXRequest: "false", htmx.HXBoosted: "false"}, false},
		{"request is nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result := htmx.IsHtmxRequest(req)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsBoostedRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{"no headers", map[string]string{}, false},
		{"HXBoosted true", map[string]string{htmx.HXBoosted: "true"}, true},
		{"HXBoosted false", map[string]string{htmx.HXBoosted: "false"}, false},
		{"request is nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result := htmx.IsBoostedRequest(req)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsAnyHtmxRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{"no headers", map[string]string{}, false},
		{"only HXRequest true", map[string]string{htmx.HXRequest: "true"}, true},
		{"only HXBoosted true", map[string]string{htmx.HXBoosted: "true"}, true},
		{"both HXRequest and HXBoosted true", map[string]string{htmx.HXRequest: "true", htmx.HXBoosted: "true"}, true},
		{"request is nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result := htmx.IsAnyHtmxRequest(req)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsHistoryRestoreRequest(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected bool
	}{
		{"no headers", map[string]string{}, false},
		{"HXHistoryRestoreRequest true", map[string]string{htmx.HXHistoryRestoreRequest: "true"}, true},
		{"HXHistoryRestoreRequest false", map[string]string{htmx.HXHistoryRestoreRequest: "false"}, false},
		{"request is nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result := htmx.IsHistoryRestoreRequest(req)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCurrentURL(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
		found    bool
	}{
		{"no headers", map[string]string{}, "", false},
		{"HXCurrentURL present", map[string]string{htmx.HXCurrentURL: "https://example.com"}, "https://example.com", true},
		{"request is nil", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result, found := htmx.CurrentURL(req)
			if result != tt.expected || found != tt.found {
				t.Errorf("expected (%v, %v), got (%v, %v)", tt.expected, tt.found, result, found)
			}
		})
	}
}

func TestPrompt(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
		found    bool
	}{
		{"no headers", map[string]string{}, "", false},
		{"HXPrompt present", map[string]string{htmx.HXPrompt: "value"}, "value", true},
		{"request is nil", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result, found := htmx.Prompt(req)
			if result != tt.expected || found != tt.found {
				t.Errorf("expected (%v, %v), got (%v, %v)", tt.expected, tt.found, result, found)
			}
		})
	}
}

func TestTarget(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
		found    bool
	}{
		{"no headers", map[string]string{}, "", false},
		{"HXTarget present", map[string]string{htmx.HXTarget: "targetId"}, "targetId", true},
		{"request is nil", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result, found := htmx.Target(req)
			if result != tt.expected || found != tt.found {
				t.Errorf("expected (%v, %v), got (%v, %v)", tt.expected, tt.found, result, found)
			}
		})
	}
}

func TestTrigger(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
		found    bool
	}{
		{"no headers", map[string]string{}, "", false},
		{"HXTrigger present", map[string]string{htmx.HXTrigger: "triggerId"}, "triggerId", true},
		{"request is nil", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result, found := htmx.Trigger(req)
			if result != tt.expected || found != tt.found {
				t.Errorf("expected (%v, %v), got (%v, %v)", tt.expected, tt.found, result, found)
			}
		})
	}
}

func TestTriggerName(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
		found    bool
	}{
		{"no headers", map[string]string{}, "", false},
		{"HXTriggerName present", map[string]string{htmx.HXTriggerName: "triggerName"}, "triggerName", true},
		{"request is nil", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := buildRequestWithHeaders(tt.headers)
			result, found := htmx.TriggerName(req)
			if result != tt.expected || found != tt.found {
				t.Errorf("expected (%v, %v), got (%v, %v)", tt.expected, tt.found, result, found)
			}
		})
	}
}

func buildRequestWithHeaders(headers map[string]string) *http.Request {
	if headers == nil {
		return nil
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	return req
}
