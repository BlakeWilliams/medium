package medium

import (
	"context"
	"net/http"
)

type dispatchable[T Action] interface {
	dispatch(r *http.Request, rw http.ResponseWriter) (bool, map[string]string, func(T))
}

// Represents the next middleware to be called in the middleware stack.
type NextMiddleware func(context.Context, *http.Request, http.ResponseWriter)

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
// type Middleware func(c Action, next HandlerFunc[Action])
type Middleware func(context.Context, *http.Request, http.ResponseWriter, NextMiddleware)

// A function that handles a request.
type HandlerFunc[T any] func(T)

// Convenience type for middleware handlers
// type MiddlewareFunc = HandlerFunc[Action]

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T Action] struct {
	routes        []*Route[T]
	middleware    []Middleware
	actionCreator func(context.Context, Action, func(T))
	// Called when no route matches the request. Useful for rendering 404 pages.
	missingRoute HandlerFunc[T]

	groups []dispatchable[T]
}

// Creates a new Router with the given action creator used to create the application's root type.
func New[T Action](actionCreator func(context.Context, Action, func(T))) *Router[T] {
	return &Router[T]{
		actionCreator: actionCreator,
		routes:        make([]*Route[T], 0),
	}
}

func (router *Router[T]) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var handler NextMiddleware

	handler = func(ctx context.Context, r *http.Request, rw http.ResponseWriter) {
		ok, params, routeHandler := router.dispatch(ctx, r, rw)

		var mediumHandler func(ctx context.Context, a Action)
		if ok {
			mediumHandler = func(ctx context.Context, action Action) {
				router.actionCreator(ctx, action, func(action T) {
					routeHandler(ctx, action)
				})
			}
		}

		if !ok {
			mediumHandler = func(ctx context.Context, action Action) {
				router.actionCreator(ctx, action, func(action T) {
					if router.missingRoute != nil {
						router.missingRoute(action)
					} else {
						action.ResponseWriter().WriteHeader(http.StatusNotFound)
						_, _ = action.Write([]byte("404 not found"))
					}
				})
			}
		}

		action := NewAction(rw, r, params)
		mediumHandler(ctx, action)
	}

	for i := len(router.middleware) - 1; i >= 0; i-- {
		middleware := router.middleware[i]
		nextHandler := handler

		handler = func(ctx context.Context, r *http.Request, rw http.ResponseWriter) {
			middleware(ctx, r, rw, nextHandler)
		}
	}

	handler(context.Background(), r, rw)

}

func (router *Router[T]) dispatch(ctx context.Context, r *http.Request, rw http.ResponseWriter) (bool, map[string]string, func(context.Context, T)) {
	if route, params := router.routeFor(r); route != nil {
		return true, params, func(ctx context.Context, action T) {
			route.handler(action)
		}
	}

	for _, group := range router.groups {
		if ok, params, handler := group.dispatch(r, rw); ok {
			return true, params, func(ctx context.Context, action T) {
				handler(action)
			}
		}
	}

	return false, nil, nil
}

func (router *Router[T]) routeFor(r *http.Request) (*Route[T], map[string]string) {
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
