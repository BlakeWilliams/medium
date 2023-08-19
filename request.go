package medium

import (
	"io"
	"net/http"
)

// ResponseWriter is an interface that represents the response that will be sent to the
// client.
type ResponseWriter interface {
	// Status returns the status code that will be written to the response.
	Status() int
	// Body returns the body that will be written to the response.
	Body() io.ReadCloser
	// Header returns the headers that will be written to the response.
	Header() http.Header
	// Writer returns the original response writer.
	Writer() http.ResponseWriter
	http.ResponseWriter
}

type response struct {
	status         int
	headers        http.Header
	body           io.ReadCloser
	responseWriter http.ResponseWriter
}

var _ ResponseWriter = (*response)(nil)

func (r response) Status() int                 { return r.status }
func (r response) Body() io.ReadCloser         { return r.body }
func (r response) Header() http.Header         { return r.headers }
func (r response) Writer() http.ResponseWriter { return r.responseWriter }
func (r response) WriteHeader(status int)      { r.responseWriter.WriteHeader(status) }
func (r response) Write(b []byte) (int, error) { return r.responseWriter.Write(b) }

type handlerContext interface {
	Request() *http.Request
	Response() ResponseWriter
}

// RootRequest is a wrapper around http.Request that contains the original Request
// object and the response writer. This is used for the root router since there is
// no application specific data to store.
type RootRequest struct {
	originalRequest *http.Request
	response        ResponseWriter
}

var _ handlerContext = (*RootRequest)(nil)

// Writer returns the original response writer.
func (r RootRequest) Writer() http.ResponseWriter { return r.response }

// Request returns the original request.
func (r RootRequest) Request() *http.Request { return r.originalRequest }

// Response returns the response that will be sent to the client.
func (r RootRequest) Response() ResponseWriter { return r.response }

// Request is a wrapper around http.Request that contains the original Request
// object and the response writer. It also can contain application specific data
// that is passed to the handlers.
type Request[Data any] struct {
	root        RootRequest
	routeParams map[string]string
	Data        Data
}

var _ handlerContext = (*Request[any])(nil)

// Writer returns the original response writer.
func (r Request[Data]) Writer() http.ResponseWriter { return r.root.response }

// Request returns the original request.
func (r Request[Data]) Request() *http.Request { return r.root.originalRequest }

// Response returns the response that will be sent to the client.
func (r Request[Data]) Response() ResponseWriter { return r.root.response }

// Params returns the route parameters that were matched.
func (r Request[Data]) Params() map[string]string { return r.routeParams }
