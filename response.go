package medium

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// ResponseWriter is an interface that represents the response that will be sent to the
// client.
//
// The default status if not provided is 200, and the default headers are an empty map.
type Response interface {
	Status() int
	Header() http.Header
	Body() io.Reader
}

type basicResponse struct {
	status int
	header http.Header
	body   io.Reader
}

func (br basicResponse) Status() int {
	if br.status == 0 {
		return http.StatusOK
	}

	return br.status
}
func (br basicResponse) Header() http.Header { return br.header }
func (br basicResponse) Body() io.Reader     { return br.body }

func FullResponse(status int, header http.Header, body io.Reader) Response {
	return basicResponse{status: status, header: header, body: body}
}

func StringResponse(status int, body string) Response {
	return basicResponse{status: status, header: http.Header{}, body: strings.NewReader(body)}
}

// OK returns a response with a 200 status code and a body of "OK".
func OK() Response {
	return StringResponse(http.StatusOK, "OK")
}

type ResponseBuilder struct {
	status int
	header http.Header
	body   io.Reader
}

func (rb *ResponseBuilder) Status(status int)      { rb.status = status }
func (rb *ResponseBuilder) Body(body io.Reader)    { rb.body = body }
func (rb *ResponseBuilder) StringBody(body string) { rb.body = strings.NewReader(body) }
func (rb *ResponseBuilder) BytesBody(body []byte)  { rb.body = bytes.NewReader(body) }
func (rb *ResponseBuilder) Header(key, value string) {
	if rb.header == nil {
		rb.header = http.Header{}
	}

	rb.header.Set(key, value)
}

func (rb ResponseBuilder) Response() Response {
	return basicResponse{status: rb.status, header: rb.header, body: rb.body}
}
