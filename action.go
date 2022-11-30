package medium

import (
	"context"
	"net/http"
	"net/url"
)

func DefaultActionFactory(ctx context.Context, rctx *RouterContext[*BaseAction]) {
	action := rctx.Action().(*BaseAction)

	rctx.Next(ctx, action)
}

// Creates a new BaseAction which implements the Action interface. It is used as
// the base for most applications and should be extended via embedding to allow
// consumers to provide their own fields and methods in struct embedding
// BaseAction.
//
// The default Context() for a BaseAction inherits the context from the *http.Request.
func NewAction(rw http.ResponseWriter, r *http.Request, params map[string]string) *BaseAction {
	baseAction := &BaseAction{
		responseWriter: rw,
		request:        r,
		params:         params,
		status:         200,
		context:        r.Context(),
	}

	newRw := &statusForwarder{
		originalResponseWriter: rw,
		onWriteHeader:          func(status int) { baseAction.status = status },
	}
	baseAction.SetResponseWriter(newRw)

	return baseAction
}

// Action is the interface that implements the basic methods for an action. It's
// intended that consumers of the package implement their own concrete Actions
// that embed `Action` and extend their struct with application/context specific
// behavior.
type Action interface {
	// Returns the original http.ResponseWriter
	ResponseWriter() http.ResponseWriter
	// Sets this actions response writer. This can be used for wrapping the
	// existing ResponseWriter with new functionality, such as capturing the
	// status code.
	SetResponseWriter(http.ResponseWriter)
	// Returns the original http.Request
	Request() *http.Request
	// Sets the request for this action
	SetRequest(*http.Request)
	// Returns the parameters derived from the route in the router. e.g.
	// `/user/:id` would make `id` available as a parameter.
	Params() map[string]string
	// Write implements the io.Writer interface and writes the given content to
	// the response writer.
	Write(content []byte) (int, error)
	// WithContext should modify the underlying *http.Request context,
	// allowing medium and non-medium context consumers to utilize context.
	WithContext(context.Context) context.Context
	// Context is a convenience function for returning Request().Context()
	Context() context.Context
}

type BaseAction struct {
	responseWriter http.ResponseWriter
	request        *http.Request
	params         map[string]string
	status         int
	context        context.Context
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

func (ba *BaseAction) SetRequest(r *http.Request) {
	ba.request = r
}

func (ba *BaseAction) Context() context.Context {
	return ba.request.Context()
}

func (ba *BaseAction) WithContext(ctx context.Context) context.Context {
	ba.request = ba.request.WithContext(ctx)
	return ba.request.Context()
}

var _ Action = &BaseAction{}
