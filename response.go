package medium

import (
	"bytes"
	"io"
	"net/http"
)

// NoData is a placeholder type for the default action creator.
type NoData struct{}

// WithNoData is a convenience function for creating a NoData type for use with
// groups and routers.
func WithNoData(rootRequest *RootRequest) NoData {
	return NoData{}
}

// ResponseWriter is an interface that represents the response that will be sent to the
// client.
//
// The default status if not provided is 200, and the default headers are an empty map.
type Response interface {
	Status() int
	Header() http.Header
	Body() io.Reader
}

// NewResponse returns a new ResponseBuilder that can be used to build a response.
func NewResponse() *ResponseBuilder {
	return &ResponseBuilder{
		status: http.StatusOK,
		header: http.Header{},
	}
}

func StringResponse(status int, body string) Response {
	res := NewResponse()
	res.WriteStatus(status)
	res.WriteString(body)

	return res
}

// OK returns a response with a 200 status code and a body of "OK".
func OK() Response {
	return StringResponse(http.StatusOK, "OK")
}

// Redirect returns an HTTP response to redirect the client to the provided URL.
func Redirect(to string) Response {
	res := NewResponse()
	res.WriteStatus(http.StatusFound)
	res.Header().Set("Location", to)
	res.WriteString("redirecting to " + to)

	return res
}

// ResponseBuilder is a helper struct that can be used to build a response. It
// implements the response interface and can be returned directly from handlers.
type ResponseBuilder struct {
	status int
	header http.Header
	body   io.Reader
}

var _ io.Writer = (*ResponseBuilder)(nil)
var _ Response = (*ResponseBuilder)(nil)

// WriteStatus sets the status code for the response. It does not prevent
// the status code from being changed by a middleware or writing additional
// headers.
func (rb *ResponseBuilder) WriteStatus(status int) { rb.status = status }

// Write writes the provided bytes to the response body.
func (rb *ResponseBuilder) Write(p []byte) (int, error) {
	if rb.body == nil {
		rb.body = bytes.NewReader(p)
	} else {
		rb.body = io.MultiReader(rb.body, bytes.NewReader(p))
	}

	return len(p), nil
}

// WriteString writes the provided string to the response body.
func (rb *ResponseBuilder) WriteString(s string) (int, error) {
	return rb.Write([]byte(s))
}

// Body returns the body of the response.
func (rb *ResponseBuilder) Body() io.Reader { return rb.body }

// Status returns the status code of the response.
func (rb *ResponseBuilder) Status() int { return rb.status }

// Header returns the header map for the response.
func (rb *ResponseBuilder) Header() http.Header {
	if rb.header == nil {
		rb.header = http.Header{}
	}

	return rb.header
}
