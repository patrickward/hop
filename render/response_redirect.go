package render

import (
	"encoding/json"
	"net/http"

	"github.com/patrickward/hop/render/htmx"
	"github.com/patrickward/hop/render/request"
)

// RedirectWithHTMX sends an HX-Redirect header to the client
func (resp *Response) RedirectWithHTMX(w http.ResponseWriter, url string) {
	w.Header().Set(htmx.HXRedirect, url)
	w.WriteHeader(http.StatusSeeOther)
	_, _ = w.Write([]byte("redirecting..."))
}

// Redirect sends a redirect response to the client
func (resp *Response) Redirect(w http.ResponseWriter, r *http.Request, url string) {
	if htmx.IsHtmxRequest(r) {
		resp.RedirectWithHTMX(w, url)
		return
	} else if request.IsXMLHttpRequest(r) {
		// Create a JSON response with a redirect
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := map[string]string{
			"status":  "redirect",
			"message": "redirecting...",
			"url":     url,
		}

		jsonBytes, _ := json.Marshal(data)
		_, _ = w.Write(jsonBytes)
		return
	}

	// Otherwise, send a standard redirect
	http.Redirect(w, r, url, http.StatusFound)
}
