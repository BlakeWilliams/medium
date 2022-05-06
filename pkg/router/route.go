package router

import (
	"net/http"
	"strings"
)

type Route[T Controller] struct {
	Method  string
	Raw     string
	parts   []string
	handler func(T)
}

func (r *Route[T]) IsMatch(req *http.Request) (bool, map[string]string) {
	if r.Method != req.Method {
		return false, nil
	}

	reqParts := strings.Split(req.URL.Path, "/")

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

func newRoute[T Controller](method string, path string, handler func(T)) *Route[T] {
	// TODO better support for `/`, remove double `//`
	return &Route[T]{Method: method, Raw: path, parts: strings.Split(path, "/"), handler: handler}
}
