package render_test

import (
	"net/http"
	"testing"

	"github.com/patrickward/hop/v2/render"
	"github.com/patrickward/hop/v2/render/htmx"
	"github.com/patrickward/hop/v2/render/htmx/location"
	"github.com/patrickward/hop/v2/render/htmx/swap"
)

func TestResponseSetsHxLocation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		path     string
		opts     []location.Option
		expected string
	}{
		{
			name:     "Simple path",
			path:     "/new-path",
			expected: "/new-path",
		},
		{
			name: "Path with innerHTML swap",
			path: "/new-path",
			opts: []location.Option{
				location.Swap(swap.InnerHTML(swap.Transition(true), swap.IgnoreTitle())),
			},
			expected: `{"path":"/new-path","swap":"innerHTML transition:true ignoreTitle:true"}`,
		},
		{
			name: "Path with outerHTML swap",
			path: "/new-path",
			opts: []location.Option{
				location.Handler("handleClick"),
				location.Headers(map[string]string{"header": "value"}),
				location.Select("#foobar"),
			},
			expected: `{"path":"/new-path","handler":"handleClick","headers":{"header":"value"},"select":"#foobar"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rs := render.NewResponse(createResponseTemplateManager(t))
			rs.HxLocation(tc.path, tc.opts...)
			td := rs.ToTemplateData(createSimpleRequest(t))

			if td == nil {
				t.Fatal("Expected TemplateData to be set, but it was nil")
			}

			if td.Headers == nil {
				t.Fatal("Expected Headers to be set, but it was nil")
			}

			if td.Headers[htmx.HXLocation] != tc.expected {
				t.Errorf("Expected HX-Location header to be set to '%s', but got '%s'", tc.expected, td.Headers[htmx.HXLocation])
			}
		})
	}
}

func TestResponseSetsHxPushURL(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxPushURL("/new-path")
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXPushURL] != "/new-path" {
		t.Errorf("Expected HX-Push-Url header to be set to '/new-path', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXPushURL])
	}
}

func TestResponseSetsNoHxPushURL(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxNoPushURL()
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXPushURL] != "false" {
		t.Errorf("Expected HX-Push-Url header to be set to 'false', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXPushURL])
	}
}

func TestResponseSetsHxRedirect(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxRedirect("/redirect-path")
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXRedirect] != "/redirect-path" {
		t.Errorf("Expected HX-Redirect header to be set to '/redirect-path', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXRedirect])
	}
}

func TestResponseSetsHxRefresh(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxRefresh()
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXRefresh] != "true" {
		t.Errorf("Expected HX-Refresh header to be set to 'true', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXRefresh])
	}
}

func TestResponseSetsNoHxRefresh(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxNoRefresh()
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXRefresh] != "false" {
		t.Errorf("Expected HX-Refresh header to be set to 'false', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXRefresh])
	}
}

func TestResponseSetsHxReplaceURL(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxReplaceURL("/replace-path")
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXReplaceURL] != "/replace-path" {
		t.Errorf("Expected HX-Replace-Url header to be set to '/replace-path', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXReplaceURL])
	}
}

func TestResponseSetsNoHxReplaceURL(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxNoReplaceURL()
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXReplaceURL] != "false" {
		t.Errorf("Expected HX-Replace-Url header to be set to 'false', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXReplaceURL])
	}
}

func TestResponseSetsHxReswap(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxReswap(swap.InnerHTML())
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXReswap] != "innerHTML" {
		t.Errorf("Expected HX-Reswap header to be set to 'innerHTML', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXReswap])
	}
}

func TestResponseSetsHxRetarget(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxRetarget("#new-target")
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXRetarget] != "#new-target" {
		t.Errorf("Expected HX-Retarget header to be set to '#new-target', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXRetarget])
	}
}

func TestResponseSetsHxReselect(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxReselect("#new-select")
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXReselect] != "#new-select" {
		t.Errorf("Expected HX-ReSelect header to be set to '#new-select', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXReselect])
	}
}

func TestResponseSetsHxTrigger(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxTrigger("test-trigger", "test-value")
	rs.HxTrigger("another-trigger", "another-value")
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXTrigger] != `{"another-trigger":"another-value","test-trigger":"test-value"}` {
		t.Errorf("Expected HX-Trigger header to include 'another-trigger', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXTrigger])
	}
}

func TestResponseSetsHxTriggerAfterSettle(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxTriggerAfterSettle("test-trigger", "test-value")
	rs.HxTriggerAfterSettle("another-trigger", "another-value")
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXTriggerAfterSettle] != `{"another-trigger":"another-value","test-trigger":"test-value"}` {
		t.Errorf("Expected HX-Trigger-After-Settle header to include 'another-trigger', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXTriggerAfterSettle])
	}
}

func TestResponseSetsHxTriggerAfterSwap(t *testing.T) {
	t.Parallel()

	rs := render.NewResponse(createResponseTemplateManager(t))
	rs.HxTriggerAfterSwap("test-trigger", "test-value")
	rs.HxTriggerAfterSwap("another-trigger", "another-value")
	req := createSimpleRequest(t)
	if rs.ToTemplateData(req).Headers[htmx.HXTriggerAfterSwap] != `{"another-trigger":"another-value","test-trigger":"test-value"}` {
		t.Errorf("Expected HX-Trigger-After-Swap header to include 'another-trigger', but got '%s'", rs.ToTemplateData(req).Headers[htmx.HXTriggerAfterSwap])
	}
}

func TestResponseReturnsAppropriateLayout(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		req            func() *http.Request
		expectedLayout string
	}{
		{
			name: "HTMX request",
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set(htmx.HXRequest, "true")
				return req
			},
			expectedLayout: "htmx",
		},
		{
			name: "Non-HTMX request",
			req: func() *http.Request {
				req, _ := http.NewRequest("GET", "/test", nil)
				return req
			},
			expectedLayout: "default",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rs := render.NewResponse(createResponseTemplateManager(t))
			rs.HxLayout(tc.req(), "htmx", "default")
			td := rs.ToTemplateData(createSimpleRequest(t))
			if td == nil {
				t.Fatal("Expected TemplateData to be set, but it was nil")
			}

			if td.Layout != tc.expectedLayout {
				t.Errorf("Expected layout to be '%s', but got '%s'", tc.expectedLayout, td.Layout)
			}
		})
	}
}
