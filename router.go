package medium

import (
	"net/http"
)

type dispatchable[T any] interface {
	dispatch(r RootRequest) (bool, map[string]string, func(Request[T]))
}

var _ dispatchable[Action] = (*Router[Action])(nil)
var _ dispatchable[Action] = (*Group[Action, Action])(nil)

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
// type Middleware func(c Action, next HandlerFunc[Action])
type Middleware func(http.ResponseWriter, *http.Request, http.HandlerFunc)

// A function that handles a request.
type HandlerFunc[T any] func(Request[T])

// Convenience type for middleware handlers
// type MiddlewareFunc = HandlerFunc[Action]

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T any] struct {
	routes      []*Route[T]
	middleware  []Middleware
	dataCreator func(RootRequest, func(T))
	// Called when no route matches the request. Useful for rendering 404 pages.
	missingRoute HandlerFunc[T]

	groups []dispatchable[T]
}

// Creates a new Router with the given action creator used to create the application's root type.
func New[T any](dataCreator func(RootRequest, func(T))) *Router[T] {
	return &Router[T]{
		dataCreator: dataCreator,
		routes:      make([]*Route[T], 0),
	}
}

func (router *Router[T]) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var handler http.HandlerFunc

	handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rootRequest := RootRequest{originalRequest: r, response: &response{responseWriter: rw}}
		ok, params, routeHandler := router.dispatch(rootRequest)

		var mediumHandler func()

		if ok {
			mediumHandler = func() {
				router.dataCreator(rootRequest, func(data T) {
					routeHandler(Request[T]{root: rootRequest, routeParams: params, Data: data})
				})
			}
		} else {
			mediumHandler = func() {
				router.dataCreator(rootRequest, func(data T) {
					if router.missingRoute != nil {
						router.missingRoute(Request[T]{root: rootRequest, routeParams: params, Data: data})
					} else {
						rootRequest.Response().WriteHeader(http.StatusNotFound)
						_, _ = rootRequest.Response().Write([]byte("404 not found"))
					}
				})
			}
		}

		mediumHandler()
	})

	for i := len(router.middleware) - 1; i >= 0; i-- {
		middleware := router.middleware[i]
		nextHandler := handler

		handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			middleware(rw, r, nextHandler)
		})
	}

	handler.ServeHTTP(rw, r)
}

func (router *Router[T]) dispatch(r RootRequest) (bool, map[string]string, func(Request[T])) {
	if route, params := router.routeFor(r); route != nil {
		return true, params, route.handler
	}

	for _, group := range router.groups {
		if ok, params, handler := group.dispatch(r); ok {
			return true, params, handler
		}
	}

	return false, nil, nil
}

func (router *Router[T]) routeFor(r RootRequest) (*Route[T], map[string]string) {
	for _, route := range router.routes {
		if ok, params := route.IsMatch(r); ok {
			return route, params
		}
	}

	return nil, nil
}

// Match is used to add a new Route to the Router
func (r *Router[T]) Match(method string, path string, handler HandlerFunc[T]) {
	r.routes = append(r.routes, newRoute(method, path, handler))
}

// Defines a new Route that responds to GET requests.
func (r *Router[T]) Get(path string, handler HandlerFunc[T]) {
	r.Match(http.MethodGet, path, handler)
}

// Defines a new Route that responds to POST requests.
func (r *Router[T]) Post(path string, handler HandlerFunc[T]) {
	r.Match(http.MethodPost, path, handler)
}

// Defines a new Route that responds to PUT requests.
func (r *Router[T]) Put(path string, handler HandlerFunc[T]) {
	r.Match(http.MethodPut, path, handler)
}

// Defines a new Route that responds to PATCH requests.
func (r *Router[T]) Patch(path string, handler HandlerFunc[T]) {
	r.Match(http.MethodPatch, path, handler)
}

// Defines a new Route that responds to DELETE requests.
func (r *Router[T]) Delete(path string, handler HandlerFunc[T]) {
	r.Match(http.MethodDelete, path, handler)
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
// actionCreator function passed to New or NewGroup.
//
// Middleware is called in the order that they are added. Middleware must call
// next in order to continue the request, otherwise the request is halted.
func (r *Router[T]) Use(middleware Middleware) {
	r.middleware = append(r.middleware, middleware)
}

var _ registerable[Action] = (*Router[Action])(nil)

func (r *Router[T]) register(group dispatchable[T]) {
	r.groups = append(r.groups, group)
}

func (r *Router[T]) prefix() string { return "" }
