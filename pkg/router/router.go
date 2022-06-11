package router

import (
	"net/http"
)

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
type Middleware func(*BaseAction, HandlerFunc[*BaseAction])

// AroundHandler represents a function that wraps a given route handler.
type AroundHandler[T Action] func(T, func())

// ContextFactory is a function that returns a new context for each request.
// This is the entrypoint for the router and can be used to setup request data
// like fetching the current user, reading session data, etc.
type ContextFactory[T any] func(*BaseAction) T

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T Action] struct {
	routes         []*Route[T]
	middleware     []Middleware
	aroundHandlers []AroundHandler[T]
	contextFactory ContextFactory[T]
	// Called when no route matches the request. Useful for rendering 404 pages.
	missingRoute HandlerFunc[T]
}

// Creates a new Router with the given ContextFactory.
func New[T Action](contextFactory ContextFactory[T]) *Router[T] {
	return &Router[T]{
		contextFactory: contextFactory,
		routes:         make([]*Route[T], 0),
	}
}

// DEPRECATED: Use ServeHttp
// Run is responsible dispatching requests to the correct handler.
// First the middleware is run, then if there is a matching route, the handler
// associated with that Route is called.
func (router *Router[T]) Run(rw http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(rw, r)
}

func (router *Router[T]) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var matchingRoute *Route[T]
	params := map[string]string{}

	for _, route := range router.routes {
		if ok, routeParams := route.IsMatch(r); ok {
			matchingRoute = route
			params = routeParams
			break
		}
	}

	var handler HandlerFunc[*BaseAction]

	// TODO - there's no reason we need to re-build the middleware stack and
	// around stack each request.
	if matchingRoute != nil {
		handler = func(baseAction *BaseAction) {
			ac := router.contextFactory(baseAction)
			wrappedHandler := router.wrapHandler((matchingRoute.handler))
			wrappedHandler(ac)
		}
	} else if router.missingRoute != nil {
		handler = func(baseAction *BaseAction) {
			baseAction.Response().WriteHeader(http.StatusNotFound)
			action := router.contextFactory(baseAction)
			router.missingRoute(action)
		}
	} else {
		handler = func(baseAction *BaseAction) {
			baseAction.Response().WriteHeader(http.StatusNotFound)
			baseAction.Write([]byte("404 not found"))
		}
	}

	action := NewAction(rw, r, params)
	next := handler

	for i := len(router.middleware) - 1; i >= 0; i-- {
		newNext := func(next HandlerFunc[*BaseAction], middleware Middleware) func(baseAction *BaseAction) {
			return func(baseAction *BaseAction) {
				middleware(baseAction, next)
			}
		}(next, router.middleware[i])
		next = newNext
	}

	next(action)
}

func (r *Router[T]) wrapHandler(baseHandler HandlerFunc[T]) HandlerFunc[T] {
	currentHandler := baseHandler

	for i := len(r.aroundHandlers) - 1; i >= 0; i-- {
		newHandler := func(handler AroundHandler[T], next HandlerFunc[T]) func(ac T) {
			return func(ac T) {
				handler(ac, func() {
					next(ac)
				})
			}
		}(r.aroundHandlers[i], currentHandler)

		currentHandler = newHandler
	}

	return currentHandler
}

// Match is used to add a new Route to the Router
func (r *Router[T]) Match(method string, path string, handler HandlerFunc[T]) {
	r.routes = append(r.routes, newRoute(method, path, handler))
}

// Defines a new Route that responds to GET requests.
func (r *Router[T]) Get(path string, handler HandlerFunc[T]) {
	r.routes = append(r.routes, newRoute(http.MethodGet, path, handler))
}

// Defines a new Route that responds to POST requests.
func (r *Router[T]) Post(path string, handler HandlerFunc[T]) {
	r.routes = append(r.routes, newRoute(http.MethodPost, path, handler))
}

// Defines a handler that is called when no route matches the request.
func (r *Router[T]) Missing(handler HandlerFunc[T]) {
	r.missingRoute = handler
}

// Defines a new middleware that is called in each request before the matching
// route is called, if one exists. Middleware are only passed a
// router.BaseAction and not the application specific action. This is due to
// middleware being treated as a low-level API.
//
// If you need access to the application specific action, you can use the
// router.Around method.
//
// Middleware is called in the order that they are added. Middleware must call
// next in order to continue the request, otherwise the request is halted.
func (r *Router[T]) Use(middleware Middleware) {
	r.middleware = append(r.middleware, middleware)
}

// Defines a new AroundHandler that is called before matching routes or
// missingRoute handlers are called.
//
// Around handlers are passed a function that should be called to continue the
// request. If it is not called, the request is halted.
func (r *Router[T]) Around(aroundHandler AroundHandler[T]) {
	r.aroundHandlers = append(r.aroundHandlers, aroundHandler)
}
