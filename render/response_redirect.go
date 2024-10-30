package render

import (
	"net/http"

	"github.com/patrickward/hop/render/htmx"
)

func (resp *Response) Redirect(w http.ResponseWriter, r *http.Request, url string, status int) {
	if htmx.IsHtmxRequest(r) {
		resp.HxRedirect(url)
		w.WriteHeader(status)
		return
	}

	http.Redirect(w, r, url, status)
}
