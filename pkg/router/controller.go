package router

import "net/http"

func DefaultControllerFactory(w http.ResponseWriter, r *http.Request, params map[string]string) BaseController {
	return BaseController{Request: r, Response: w, params: params}
}

type Controller interface {
	Write([]byte) (int, error)
	Params() map[string]string
}

type BaseController struct {
	Response http.ResponseWriter
	Request  *http.Request
	params   map[string]string
}

func (bc BaseController) Write(content []byte) (int, error) {
	return bc.Response.Write(content)
}

func (bc BaseController) Params() map[string]string {
	return bc.params
}

var _ Controller = (*BaseController)(nil)
