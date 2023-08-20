package medium

import (
	"net/http"
	"strings"
)

type requestable interface {
	Request() *http.Request
}

// A Route is a single route that can be matched against a request and holds a
// reference to the handler used to handle the reques and holds a reference to
// the handler used to handle the request.
type Route[C any] struct {
	Method  string
	Raw     string
	parts   []string
	handler HandlerFunc[C]
}

// Given a request, returns true if the route matches the request and false if
// not.
func (r *Route[C]) IsMatch(req *RootRequest) (bool, map[string]string) {
	if r.Method != req.Request().Method {
		return false, nil
	}

	reqParts := strings.Split(req.Request().URL.Path, "/")

	if len(r.parts) != len(reqParts) {
		return false, nil
	}

	params := make(map[string]string)

	for i, part := range r.parts {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = reqParts[i]
		} else if part != reqParts[i] {
			return false, nil
		}
	}

	return true, params
}

func newRoute[C any](method string, path string, handler HandlerFunc[C]) *Route[C] {
	// TODO better support for `/`, remove double `//`
	return &Route[C]{Method: method, Raw: path, parts: strings.Split(path, "/"), handler: handler}
}
