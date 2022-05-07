package router

import (
	"net/http"
)

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
type Middleware[T any] func(T, HandlerFunc[T])

// ContextFactory is a function that returns a new context for each request.
// This is the entrypoint for the router and can be used to setup request data
// like fetching the current user, reading session data, etc.
type ContextFactory[T any] func(http.ResponseWriter, *http.Request, map[string]string) T

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T Action] struct {
	routes         []*Route[T]
	middleware     []Middleware[T]
	contextFactory ContextFactory[T]
}

// Creates a new Router with the given ContextFactory.
func New[T Action](contextFactory ContextFactory[T]) *Router[T] {
	return &Router[T]{
		contextFactory: contextFactory,
		routes:         make([]*Route[T], 0),
	}
}

// Run is responsible dispatching requests to the correct handler.
// First the middleware is run, then if there is a matching route, the handler
// associated with that Route is called.
func (router *Router[T]) Run(rw http.ResponseWriter, r *http.Request) {
	var matchingRoute *Route[T]
	params := map[string]string{}

	for _, route := range router.routes {
		if ok, routeParams := route.IsMatch(r); ok {
			matchingRoute = route
			params = routeParams
			break
		}
	}

	context := router.contextFactory(rw, r, params)

	var handler HandlerFunc[T]

	if matchingRoute != nil {
		handler = func(c T) { matchingRoute.handler(c) }
	} else {
		handler = func(c T) {}
	}

	next := handler

	for i := len(router.middleware) - 1; i >= 0; i-- {
		newNext := func(next HandlerFunc[T], i int) HandlerFunc[T] {
			return func(ctx T) {
				router.middleware[i](ctx, next)
			}
		}(next, i)
		next = newNext
	}

	next(context)
}

// Match is used to add a new Route to the Router
func (r *Router[T]) Match(method string, path string, handler HandlerFunc[T]) {
	r.routes = append(r.routes, newRoute(method, path, handler))
}

// Defines a new Route that responds to GET requests.
func (r *Router[T]) Get(path string, handler HandlerFunc[T]) {
	r.routes = append(r.routes, newRoute(http.MethodGet, path, handler))
}

// Defines a new middleware that is called in each request before the matching
// route is called, if one exists.
//
// Middleware is called in the order that they are added. Middleware must call
// next in order to continue the request, otherwise the request is halted.
func (r *Router[T]) Use(middleware func(c T, next HandlerFunc[T])) {
	r.middleware = append(r.middleware, middleware)
}
