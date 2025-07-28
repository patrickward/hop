package render

import (
	"net/http"

	"github.com/patrickward/hop/v2/render/htmx"
	"github.com/patrickward/hop/v2/render/htmx/location"
	"github.com/patrickward/hop/v2/render/htmx/swap"
)

// HxLocation sets the HX-Location header, which instructs the browser to navigate to the given path without reloading the page.
//
// For simple navigation, use a path and no HxLocation options.
// For more complex navigations, use the HxLocation options to fine-tune the navigation.
//
// Sets the HX-Location header with the given path.
//
// For more information, see: https://htmx.org/headers/hx-location
func (rs *Response) HxLocation(path string, opt ...location.Option) *Response {
	if opt == nil {
		rs.headers[htmx.HXLocation] = path
	} else {
		loc := location.NewLocation(path, opt...)
		rs.headers[htmx.HXLocation] = loc.String()
	}
	return rs
}

// HxPushURL sets the HX-Push-Url header, which instructs the browser to navigate to the given path without reloading the page.
//
// To prevent the browser from updating the page, set the HX-Push-Url header to an empty string or "false".
//
// For more information, see: https://htmx.org/headers/hx-push-url
func (rs *Response) HxPushURL(path string) *Response {
	rs.headers[htmx.HXPushURL] = path
	return rs
}

// HxNoPushURL prevents the browser from updating the history stack by setting the HX-Push-Url header to "false".
//
// For more information, see: https://htmx.org/headers/hx-no-push-url
func (rs *Response) HxNoPushURL() *Response {
	rs.headers[htmx.HXPushURL] = "false"
	return rs
}

// HxRedirect sets the HX-Redirect header, which instructs the browser to navigate to the given path (this will reload the page).
//
// For more information, see: https://htmx.org/reference/#response_headers
func (rs *Response) HxRedirect(path string) *Response {
	rs.headers[htmx.HXRedirect] = path
	return rs
}

// HxRefresh sets the HX-Refresh header, which instructs the browser to reload the page.
//
// For more information, see: https://htmx.org/reference/#response_headers
func (rs *Response) HxRefresh() *Response {
	rs.headers[htmx.HXRefresh] = "true"
	return rs
}

// HxNoRefresh prevents the browser from reloading the page by setting the HX-Refresh header to "false".
//
// For more information, see: https://htmx.org/reference/#response_headers
func (rs *Response) HxNoRefresh() *Response {
	rs.headers[htmx.HXRefresh] = "false"
	return rs
}

// HxReplaceURL sets the HX-Replace-Url header, which instructs the browser to replace the history stack with the given path.
//
// For more information, see: https://htmx.org/headers/hx-replace-url
func (rs *Response) HxReplaceURL(path string) *Response {
	rs.headers[htmx.HXReplaceURL] = path
	return rs
}

// HxNoReplaceURL prevents the browser from updating the history stack by setting the HX-Replace-Url header to "false".
//
// For more information, see: https://htmx.org/headers/hx-replace-url
func (rs *Response) HxNoReplaceURL() *Response {
	rs.headers[htmx.HXReplaceURL] = "false"
	return rs
}

// HxReswap sets the HX-Reswap header, which instructs HTMX to change the swap behavior of the target element.
//
// For more information, see: https://htmx.org/attributes/hx-swap
func (rs *Response) HxReswap(swap *swap.Style) *Response {
	rs.headers[htmx.HXReswap] = swap.String()
	return rs
}

// HxRetarget sets the HX-Retarget header, which instructs HTMX to update the target element.
//
// For more information, see: https://htmx.org/reference/#response_headers
func (rs *Response) HxRetarget(target string) *Response {
	rs.headers[htmx.HXRetarget] = target
	return rs
}

// HxReselect sets the HX-ReSelect header, which instructs HTMX to update which part of the response is selected.
//
// For more information, see: https://htmx.org/reference/#response_headers and https://htmx.org/attributes/hx-select
func (rs *Response) HxReselect(reselect string) *Response {
	rs.headers[htmx.HXReselect] = reselect
	return rs
}

// HxTrigger sets a HX-Trigger header
//
// For more information, see: https://htmx.org/headers/hx-trigger/
func (rs *Response) HxTrigger(event string, value any) *Response {
	rs.triggers.Set(event, value)
	return rs
}

// HxTriggerAfterSettle sets a HX-Trigger-After-Settle header
//
// For more information, see: https://htmx.org/headers/hx-trigger/
func (rs *Response) HxTriggerAfterSettle(event string, value any) *Response {
	rs.triggers.SetAfterSettle(event, value)
	return rs
}

// HxTriggerAfterSwap sets a HX-Trigger-After-Swap header
//
// For more information, see: https://htmx.org/headers/hx-trigger/
func (rs *Response) HxTriggerAfterSwap(event string, value any) *Response {
	rs.triggers.SetAfterSwap(event, value)
	return rs
}

// HxLayout sets the layout for HTMX requests if the request is an HTMX request, otherwise it uses the default layout.
//
// Parameters:
//   - request is used to determine if the request is an HTMX request.
//   - hxLayout is the layout to use for HTMX requests.
//   - layout is the default layout to use if the request is not an HTMX request.
func (rs *Response) HxLayout(r *http.Request, hxLayout, layout string) *Response {
	if htmx.IsHtmxRequest(r) {
		rs.Layout(hxLayout)
	} else {
		rs.Layout(layout)
	}
	return rs
}
