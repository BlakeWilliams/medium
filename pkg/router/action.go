package router

import (
	"net/http"
	"net/url"
)

func DefaultActionFactory(baseAction Action) *BaseAction {
	return baseAction.(*BaseAction)
}

func NewAction(rw http.ResponseWriter, r *http.Request, params map[string]string) *BaseAction {
	return &BaseAction{
		responseWriter: rw,
		request:        r,
		params:         params,
	}
}

type Action interface {
	Redirect(path string)
	// deprecated, use ResponseWriter
	Response() http.ResponseWriter
	ResponseWriter() http.ResponseWriter
	Request() *http.Request
	Params() map[string]string
	URL() *url.URL
	Write(content []byte) (int, error)
}

type BaseAction struct {
	responseWriter http.ResponseWriter
	request        *http.Request
	params         map[string]string
}

func (bc BaseAction) Write(content []byte) (int, error) {
	return bc.Response().Write(content)
}

// deprecated, use ResponseWriter
func (bc BaseAction) Response() http.ResponseWriter {
	return bc.responseWriter
}

func (bc BaseAction) ResponseWriter() http.ResponseWriter {
	return bc.responseWriter
}

func (bc BaseAction) Request() *http.Request {
	return bc.request
}

func (bc BaseAction) Params() map[string]string {
	return bc.params
}

func (bc BaseAction) URL() *url.URL {
	return bc.request.URL
}

// Redirects the user to the given path.
func (bc BaseAction) Redirect(path string) {
	http.Redirect(bc.Response(), bc.Request(), path, http.StatusFound)
}

var _ Action = &BaseAction{}
