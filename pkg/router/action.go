package router

import "net/http"

func NewAction(rw http.ResponseWriter, r *http.Request, params map[string]string) *BaseAction {
	return &BaseAction{
		response: rw,
		request:  r,
		params:   params,
	}
}

type Action interface {
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

var _ Action = &BaseAction{}