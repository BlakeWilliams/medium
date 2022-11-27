package medium

import (
	"context"
	"net/http"
)

type dispatchable[T Action] interface {
	dispatch(ctx context.Context, r *http.Request, rw http.ResponseWriter) (bool, map[string]string, func(context.Context, T))
}

// Represents the next middleware to be called in the middleware stack.
type NextMiddleware func(context.Context, *http.Request, http.ResponseWriter)

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
// type Middleware func(c Action, next HandlerFunc[Action])
type Middleware func(context.Context, *http.Request, http.ResponseWriter, NextMiddleware)

// A function that handles a request.
type HandlerFunc[T any] func(context.Context, T)

// Convenience type for middleware handlers
// type MiddlewareFunc = HandlerFunc[Action]

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T Action] struct {
	routes        []*Route[T]
	middleware    []Middleware
	actionFactory func(context.Context, Migrator[Action, T])
	// Called when no route matches the request. Useful for rendering 404 pages.
	missingRoute HandlerFunc[T]

	groups []dispatchable[T]
}

type Migrator[T Action, N Action] interface {
	Action() T
	Next(context.Context, N)
}

type routerMigrator[T Action, N Action] struct {
	action T
	next   func(context.Context, N)
}

func (rm *routerMigrator[T, N]) Action() T {
	return rm.action
}

func (rm *routerMigrator[T, N]) Next(ctx context.Context, newAction N) {
	rm.next(ctx, newAction)
}

// Creates a new Router with the given ContextFactory.
func New[T Action](actionFactory func(context.Context, Migrator[Action, T])) *Router[T] {
	return &Router[T]{
		actionFactory: actionFactory,
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
				rm := &routerMigrator[Action, T]{
					action: action,
					next:   routeHandler,
				}

				router.actionFactory(ctx, rm)
			}
		}

		if !ok {
			mediumHandler = func(ctx context.Context, action Action) {
				rm := &routerMigrator[Action, T]{
					action: action,
				}

				if router.missingRoute != nil {
					rm.next = router.missingRoute
				} else {
					rm.next = func(ctx context.Context, a T) {
						action.ResponseWriter().WriteHeader(http.StatusNotFound)
						_, _ = action.Write([]byte("404 not found"))
					}
				}

				router.actionFactory(ctx, rm)
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
			rm := &routerMigrator[Action, T]{
				action: action,
				next:   route.handler,
			}

			router.actionFactory(ctx, rm)
		}
	}

	for _, group := range router.groups {
		if ok, params, handler := group.dispatch(ctx, r, rw); ok {
			return true, params, func(ctx context.Context, action T) {
				rm := &routerMigrator[Action, T]{
					action: action,
					next:   handler,
				}

				router.actionFactory(ctx, rm)
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
// routerFactory function passed to New or NewGroup.
//
// Middleware is called in the order that they are added. Middleware must call
// next in order to continue the request, otherwise the request is halted.
func (r *Router[T]) Use(middleware Middleware) {
	r.middleware = append(r.middleware, middleware)
}

func (r *Router[T]) register(group dispatchable[T]) {
	r.groups = append(r.groups, group)
}

func (r *Router[T]) prefix() string { return "" }

var _ registerable[Action] = (*Router[Action])(nil)
