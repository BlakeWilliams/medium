package medium

import (
	"context"
	"net/http"
	"net/url"
)

func DefaultActionFactory(ctx context.Context, baseAction Action, next func(context.Context, *BaseAction)) {
	action := baseAction.(*BaseAction)

	next(ctx, action)
}

func NewAction(rw http.ResponseWriter, r *http.Request, params map[string]string) *BaseAction {
	baseAction := &BaseAction{
		responseWriter: rw,
		request:        r,
		params:         params,
		status:         200,
	}

	newRw := &statusForwarder{
		originalResponseWriter: rw,
		onWriteHeader:          func(status int) { baseAction.status = status },
	}
	baseAction.SetResponseWriter(newRw)

	return baseAction
}

type Action interface {
	Redirect(path string)
	// deprecated, use ResponseWriter
	Response() http.ResponseWriter
	ResponseWriter() http.ResponseWriter
	// Sets this actions response writer. This can be used for wrapping the
	// existing ResponseWriter with new functionality, such as capturing the
	// status code.
	SetResponseWriter(http.ResponseWriter)
	Request() *http.Request
	Params() map[string]string
	Status() int
	URL() *url.URL
	Write(content []byte) (int, error)
}

type BaseAction struct {
	responseWriter http.ResponseWriter
	request        *http.Request
	params         map[string]string
	status         int
}

func (ba *BaseAction) Write(content []byte) (int, error) {
	return ba.ResponseWriter().Write(content)
}

// deprecated, use ResponseWriter
func (ba *BaseAction) Response() http.ResponseWriter {
	return ba.responseWriter
}

func (ba *BaseAction) ResponseWriter() http.ResponseWriter {
	return ba.responseWriter
}

func (ba *BaseAction) Request() *http.Request {
	return ba.request
}

func (ba *BaseAction) Params() map[string]string {
	return ba.params
}

func (ba *BaseAction) URL() *url.URL {
	return ba.request.URL
}

func (ba *BaseAction) Status() int {
	return ba.status
}

// Redirects the user to the given path.
func (ba *BaseAction) Redirect(path string) {
	http.Redirect(ba.Response(), ba.Request(), path, http.StatusFound)
}

func (ba *BaseAction) SetResponseWriter(rw http.ResponseWriter) {
	ba.responseWriter = rw
}

var _ Action = &BaseAction{}
