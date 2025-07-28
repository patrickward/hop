package render

import "net/http"

// Status sets the status code for the response
func (rs *Response) Status(status int) *Response {
	rs.status = status
	return rs
}

// StatusOK sets the status code to 200 OK
func (rs *Response) StatusOK() *Response {
	rs.status = http.StatusOK
	return rs
}

// StatusCreated sets the status code to 201 Created
func (rs *Response) StatusCreated() *Response {
	rs.status = http.StatusCreated
	return rs
}

// StatusAccepted sets the status code to 202 Accepted
func (rs *Response) StatusAccepted() *Response {
	rs.status = http.StatusAccepted
	return rs
}

// StatusNoContent sets the status code to 204 No Content
func (rs *Response) StatusNoContent() *Response {
	rs.status = http.StatusNoContent
	return rs
}

// StatusNotFound sets the status code to 404 Not Found
func (rs *Response) StatusNotFound() *Response {
	rs.status = http.StatusNotFound
	return rs
}

// StatusForbidden sets the status code to 403 Forbidden
func (rs *Response) StatusForbidden() *Response {
	rs.status = http.StatusForbidden
	return rs
}

// StatusUnauthorized sets the status code to 401 Unauthorized
func (rs *Response) StatusUnauthorized() *Response {
	rs.status = http.StatusUnauthorized
	return rs
}

// StatusUnavailable sets the status code to 503 Service Unavailable
func (rs *Response) StatusUnavailable() *Response {
	rs.status = http.StatusServiceUnavailable
	return rs
}

// StatusError sets the status code to 500 Internal Server Error
func (rs *Response) StatusError() *Response {
	rs.status = http.StatusInternalServerError
	return rs
}

// StatusStopPolling sets the status code to 286 and returns the Response object.
// This is useful when working HTMX and polling. Responding with a status of 286 will tell HTMX to stop polling.
// SEE: https://htmx.org/docs/#polling
func (rs *Response) StatusStopPolling() *Response {
	rs.status = 286
	return rs
}

// StatusUnprocessable sets the status code to 422 Unprocessable Entity
func (rs *Response) StatusUnprocessable() *Response {
	rs.status = http.StatusUnprocessableEntity
	return rs
}
