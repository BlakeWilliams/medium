package router

import "net/http"

func DefaultActionFactory(baseAction *BaseAction) *BaseAction {
	return baseAction
}

func NewAction(rw http.ResponseWriter, r *http.Request, params map[string]string) *BaseAction {
	return &BaseAction{
		response: rw,
		request:  r,
		params:   params,
	}
}

type Action interface {
	Redirect(path string)
	Response() http.ResponseWriter
	Request() *http.Request
	Params() map[string]string
	Write(content []byte) (int, error)
}

type BaseAction struct {
	response http.ResponseWriter
	request  *http.Request
	params   map[string]string
}

func (bc BaseAction) Write(content []byte) (int, error) {
	return bc.Response().Write(content)
}

func (bc BaseAction) Response() http.ResponseWriter {
	return bc.response
}

func (bc BaseAction) Request() *http.Request {
	return bc.request
}

func (bc BaseAction) Params() map[string]string {
	return bc.params
}

// Redirects the user to the given path.
func (bc BaseAction) Redirect(path string) {
	http.Redirect(bc.Response(), bc.Request(), path, http.StatusFound)
}

var _ Action = &BaseAction{}
