package view

import "net/http"

// ResponseWriter interface defines the basic contract for all responses
type ResponseWriter interface {
	Write(w http.ResponseWriter, r *http.Request) error
}

// ResponseExtension interface defines the contract for response extensions
type ResponseExtension interface {
	Apply(w http.ResponseWriter, r *http.Request) error
}
