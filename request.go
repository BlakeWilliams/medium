package medium

import (
	"io"
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
	root      RootRequest
	routeData *RouteData
	Data      Data
}

// RouteData holds data about the matched route.
type RouteData struct {
	// Params holds the route parameters that were matched.
	Params map[string]string
	// HandlerPath holds the path that was matched.
	HandlerPath string
}

func NewRequest[Data any](root RootRequest, data Data, routeData *RouteData) Request[Data] {
	return Request[Data]{root: root, routeData: routeData, Data: data}
}

// Request returns the original request.
func (r Request[Data]) Request() *http.Request { return r.root.originalRequest }

// Params returns the route parameters that were matched.
func (r Request[Data]) Params() map[string]string { return r.routeData.Params }

// MatchedPath returns the route path pattern that was matched.
func (r Request[Data]) MatchedPath() string { return r.routeData.HandlerPath }

// URL returns the URL of the request.
func (r Request[Data]) URL() string { return r.Request().URL.String() }

// Cookie returns the named cookie provided in the request.
func (r Request[Data]) Cookie(name string) (*http.Cookie, error) { return r.Request().Cookie(name) }

// Cookies parses and returns the HTTP cookies sent with the request.
func (r Request[Data]) Cookies() []*http.Cookie { return r.Request().Cookies() }

// Header returns the header field
func (r Request[Data]) Header() http.Header { return r.Request().Header }

// Method returns the HTTP method of the request.
func (r Request[Data]) Method() string { return r.Request().Method }

// Host returns the host of the request.
func (r Request[Data]) Host() string { return r.Request().Host }

// Proto returns the HTTP protocol version of the request.
func (r Request[Data]) Proto() string { return r.Request().Proto }

// ProtoMajor returns the HTTP protocol major version of the request.
func (r Request[Data]) ProtoMajor() int { return r.Request().ProtoMajor }

// ProtoMinor returns the HTTP protocol minor version of the request.
func (r Request[Data]) ProtoMinor() int { return r.Request().ProtoMinor }

// RemoteAddr returns the remote address of the request.
func (r Request[Data]) RemoteAddr() string { return r.Request().RemoteAddr }

// RequestURI returns the unmodified request-target of the request.
func (r Request[Data]) RequestURI() string { return r.Request().RequestURI }

// Body returns the request body.
func (r Request[Data]) Body() io.ReadCloser { return r.Request().Body }

// ContentLength returns the length of the request body.
func (r Request[Data]) ContentLength() int64 { return r.Request().ContentLength }

// FormData returns the parsed form data from the request body.
func (r Request[Data]) FormData() (map[string][]string, error) {
	err := r.Request().ParseForm()

	return r.Request().Form, err
}

// PostFormData returns the parsed form data from the request body.
func (r Request[Data]) PostFormData() (map[string][]string, error) {
	err := r.Request().ParseForm()
	return r.Request().PostForm, err
}

// Query returns the parsed query string from the request URL.
func (r Request[Data]) Query() map[string][]string { return r.Request().URL.Query() }

// QueryParam returns the first value for the named component of the query.
func (r Request[Data]) QueryParam(name string) string { return r.Request().URL.Query().Get(name) }

// QueryParams returns the values for the named component of the query.
func (r Request[Data]) QueryParams(name string) []string { return r.Request().URL.Query()[name] }
