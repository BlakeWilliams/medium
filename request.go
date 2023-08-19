package medium

import (
	"net/http"
)

// RootRequest is a wrapper around http.Request that contains the original Request
// object and the response writer. This is used for the root router since there is
// no application specific data to store.
type RootRequest struct {
	originalRequest *http.Request
}

// Request returns the original request.
func (r RootRequest) Request() *http.Request { return r.originalRequest }

// Request is a wrapper around http.Request that contains the original Request
// object and the response writer. It also can contain application specific data
// that is passed to the handlers.
type Request[Data any] struct {
	root        RootRequest
	routeParams map[string]string
	Data        Data
}

// Request returns the original request.
func (r Request[Data]) Request() *http.Request { return r.root.originalRequest }

// Params returns the route parameters that were matched.
func (r Request[Data]) Params() map[string]string { return r.routeParams }
