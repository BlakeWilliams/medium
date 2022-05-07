package router

import "net/http"

func NewContext(rw http.ResponseWriter, r *http.Request, params map[string]string) *BaseContext {
	return &BaseContext{
		response: rw,
		request:  r,
		params:   params,
	}
}

type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request
	Params() map[string]string
	Write(content []byte) (int, error)
}

type BaseContext struct {
	response http.ResponseWriter
	request  *http.Request
	params   map[string]string
}

func (bc BaseContext) Write(content []byte) (int, error) {
	return bc.Response().Write(content)
}

func (bc BaseContext) Response() http.ResponseWriter {
	return bc.response
}

func (bc BaseContext) Request() *http.Request{
	return bc.request
}

func (bc BaseContext) Params() map[string]string {
	return bc.params
}

var _ Context = &BaseContext{}
