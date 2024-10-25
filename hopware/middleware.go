package hopware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/justinas/nosurf"
)

// Middleware is a type that represents a middleware function.
type Middleware = func(next http.Handler) http.Handler

// RecoverPanic recovers from a panic and renders a system error page.
// The render function is passed as an argument to the middleware and used to render the error page.
func RecoverPanic(render func(http.ResponseWriter, *http.Request, error)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := recover()
				if err != nil {
					render(w, r, fmt.Errorf("%s", err))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders sets security headers in the HTTP response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Referrer-Policy restricts the amount of information sent in the Referer header.
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")

		// X-Content-Type-Options stops a browser from trying to MIME-sniff the content type.
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options provides clickjacking protection.
		w.Header().Set("X-Frame-Options", "deny")
		next.ServeHTTP(w, r)
	})
}

// ContentSecurityPolicyOptions contains the options for the Content-Security-Policy header.
type ContentSecurityPolicyOptions struct {
	// ChildSrc sets the sources that can be used in child contexts.
	ChildSrc string
	// ConnectSrc sets the sources that can be used for WebSockets, EventSource, and other interfaces.
	ConnectSrc string
	// DefaultSrc sets the default sources for fetch, worker, frame, embed, and object.
	DefaultSrc string
	// FontSrc sets the sources for fonts.
	FontSrc string
	// FormAction sets the sources that can be used as the target of form submissions.
	FrameSrc string
	// ImgSrc sets the sources for images.
	ImgSrc string
	// ManifestSrc sets the sources for web app manifests.
	ManifestSrc string
	// MediaSrc sets the sources for audio and video.
	MediaSrc string
	// ObjectSrc sets the sources for objects.
	ObjectSrc string
	// ScriptSrc sets the sources for scripts.
	ScriptSrc string
	// ScriptSrcElem sets the sources for inline scripts.
	ScriptSrcElem string
	// ScriptSrcAttr sets the sources for script attributes.
	ScriptSrcAttr string
	// StyleSrc sets the sources for stylesheets.
	StyleSrc string
	// StyleSrcElem sets the sources for inline styles.
	StyleSrcElem string
	// StyleSrcAttr sets the sources for style attributes.
	StyleSrcAttr string
	// WorkerSrc sets the sources for workers.
	WorkerSrc string
	// BaseURI sets the sources for the document base URL.
	BaseURI string
	// Sandbox sets the restrictions for content in an iframe.
	Sandbox string
	// FormAction sets the sources that can be used as the target of form submissions.
	FormAction string
	// FrameAncestors sets the sources that can embed the page in a frame.
	FrameAncestors string
	// ReportURI sets the URI to send reports of policy violations.
	ReportTo string
}

// ContentSecurityPolicy sets the Content-Security-Policy header to protect against XSS attacks.
func ContentSecurityPolicy(optsFunc func(opts *ContentSecurityPolicyOptions)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			options := ContentSecurityPolicyOptions{
				DefaultSrc: "'none'",
				FontSrc:    "'self'",
				ImgSrc:     "'self'",
				ScriptSrc:  "'self'",
				StyleSrc:   "'self'",
			}

			if optsFunc != nil {
				optsFunc(&options)
			}

			var v string
			v += maybeAddDirective("child-src", options.ChildSrc)
			v += maybeAddDirective("connect-src", options.ConnectSrc)
			v += maybeAddDirective("default-src", options.DefaultSrc)
			v += maybeAddDirective("font-src", options.FontSrc)
			v += maybeAddDirective("frame-src", options.FrameSrc)
			v += maybeAddDirective("img-src", options.ImgSrc)
			v += maybeAddDirective("manifest-src", options.ManifestSrc)
			v += maybeAddDirective("media-src", options.MediaSrc)
			v += maybeAddDirective("object-src", options.ObjectSrc)
			v += maybeAddDirective("script-src", options.ScriptSrc)
			v += maybeAddDirective("script-src-elem", options.ScriptSrcElem)
			v += maybeAddDirective("script-src-attr", options.ScriptSrcAttr)
			v += maybeAddDirective("style-src", options.StyleSrc)
			v += maybeAddDirective("style-src-elem", options.StyleSrcElem)
			v += maybeAddDirective("style-src-attr", options.StyleSrcAttr)
			v += maybeAddDirective("worker-src", options.WorkerSrc)
			v += maybeAddDirective("base-uri", options.BaseURI)
			v += maybeAddDirective("sandbox", options.Sandbox)
			v += maybeAddDirective("form-action", options.FormAction)
			v += maybeAddDirective("frame-ancestors", options.FrameAncestors)
			v += maybeAddDirective("report-to", options.ReportTo)

			w.Header().Set("Content-Security-Policy", strings.TrimSuffix(strings.TrimSpace(v), ";"))
			next.ServeHTTP(w, r)
		})
	}
}

func maybeAddDirective(directive, value string) string {
	if value == "" {
		return ""
	}

	return fmt.Sprintf("%s %s;", directive, value)
}

// PreventCSRF prevents CSRF attacks by setting a CSRF cookie.
func PreventCSRF(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})

	return csrfHandler
}
