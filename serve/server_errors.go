package serve

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// ReportServerError logs the error and sends an email to the admin
func (s *Server) ReportServerError(r *http.Request, err error) {
	var (
		message = err.Error()
		method  = r.Method
		url     = r.URL.String()
		trace   = string(debug.Stack())
	)

	requestAttrs := slog.Group("request", "method", method, "url", url)
	s.logger.Error(message, requestAttrs, "trace", trace)

	//if s.config.Notifications.AdminEmail != "" {
	//	data := s.NewEmailData()
	//	data["Message"] = message
	//	data["RequestMethod"] = method
	//	data["RequestURL"] = url
	//	data["Trace"] = trace
	//
	//	err := s.mailer.Send(s.config.Notifications.AdminEmail, data, "error-notification.tmpl")
	//	if err != nil {
	//		trace = string(debug.Stack())
	//		s.logger.Error(err.Error(), requestAttrs, "trace", trace)
	//	}
	//}
}
