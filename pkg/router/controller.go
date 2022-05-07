package router

import (
	"net/http"
	"strings"
)

// BeforeHandlers are defined in controllers and are run before all HTTP
// handlers defined on that controller.
//
// The BeforeHandler can prevent the execution of the HTTP handler by not
// calling the `next` function that is passed in.
type BeforeHandler[C any] func(c C, next HandlerFunc[C])

// Used to declare routes on the router without having to turn Controller into
// Controller
type routable[T any] interface {
	match(method string, path string, handler HandlerFunc[T])
}

// Controllers represent a group of related HTTP handlers. They allow you to
// declare routes, define callbacks to run before all HTTP handlers, and
// isolated shared functions within the scope they were defined in.
type Controller[C any] struct {
	rootPath       string
	beforeHandlers []BeforeHandler[C]
	router         routable[C]
}

func newController[C any](rootPath string, router routable[C]) *Controller[C] {
	return &Controller[C]{
		rootPath:       rootPath,
		beforeHandlers: make([]BeforeHandler[C], 0),
		router:         router,
	}
}

// Adds a new BeforeHandler that is run before each HTTP handler defined on
// this controller.
//
// Order matters, and Before actions will only run for handlers declared
// after it.
func (c *Controller[C]) Before(handler BeforeHandler[C]) {
	c.beforeHandlers = append(c.beforeHandlers, handler)
}

// Get declares a new route that responds to Get requests.
//
// This method accepts path argument that can be a formatted to accept
// arguments. See the Match method for more details.
func (c *Controller[C]) Get(path string, handler HandlerFunc[C]) {
	combinedPath := strings.TrimRight(c.rootPath, "/") + "/" + strings.TrimLeft(path, "/")
	beforeHandlerLen := len(c.beforeHandlers)

	// If there are no before handlers declared, run the handler directly
	if beforeHandlerLen == 0 {
		c.router.match(http.MethodGet, combinedPath, handler)
		return
	}

	next := handler

	for i := len(c.beforeHandlers) - 1; i >= 0; i-- {

		newNext := wrapHandler(c.beforeHandlers[i], next)
		next = newNext
	}

	c.router.match(http.MethodGet, combinedPath, next)
}


func wrapHandler[C any](handler BeforeHandler[C], next HandlerFunc[C]) HandlerFunc[C]{
	return func(c C) {
		handler(c, next)
	}
}
