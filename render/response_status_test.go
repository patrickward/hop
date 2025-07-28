package render_test

import (
	"net/http"
	"testing"

	"github.com/patrickward/hop/v2/render"
)

func TestResponseSetsStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		fn     func(*render.Response)
	}{
		{"Set Status directly", 404, func(rs *render.Response) { rs.Status(404) }},
		{"StatusOK", 200, func(rs *render.Response) { rs.StatusOK() }},
		{"StatusCreated", 201, func(rs *render.Response) { rs.StatusCreated() }},
		{"StatusAccepted", 202, func(rs *render.Response) { rs.StatusAccepted() }},
		{"StatusNoContent", 204, func(rs *render.Response) { rs.StatusNoContent() }},
		{"StatusNotFound", 404, func(rs *render.Response) { rs.StatusNotFound() }},
		{"StatusForbidden", 403, func(rs *render.Response) { rs.StatusForbidden() }},
		{"StatusUnauthorized", 401, func(rs *render.Response) { rs.StatusUnauthorized() }},
		{"StatusUnavailable", 503, func(rs *render.Response) { rs.StatusUnavailable() }},
		{"StatusError", 500, func(rs *render.Response) { rs.StatusError() }},
		{"StatusStopPolling", 286, func(rs *render.Response) { rs.StatusStopPolling() }},
		{"StatusUnprocessable", 422, func(rs *render.Response) { rs.StatusUnprocessable() }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest("GET", "https://example.com", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rs := render.NewResponse(createResponseTemplateManager(t))
			tc.fn(rs)
			td := rs.ToTemplateData(req)

			if td.Status != tc.status {
				t.Errorf("Expected status %d, got %d", tc.status, td.Status)
			}
		})
	}
}
