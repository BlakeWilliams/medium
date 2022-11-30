package medium

import (
	"context"
	"net/http"
)

type dispatchable[T Action] interface {
	dispatch(ctx context.Context, r *http.Request) (bool, map[string]string, func(context.Context, T))
}

// A function that handles a request.
type HandlerFunc[T any] func(context.Context, T)

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T Action] struct {
	routes        []*Route[T]
	middleware    []MiddlewareFunc
	actionFactory func(context.Context, *RouterContext[T])
	// Called when no route matches the request. Useful for rendering 404 pages.
	missingRoute HandlerFunc[T]

	groups []dispatchable[T]
}

type RouterContext[T Action] struct {
	action Action
	next   func(context.Context, T)
}

func (rc *RouterContext[T]) Action() Action {
	return rc.action
}

func (rc *RouterContext[T]) Next(ctx context.Context, nextAction T) {
	rc.next(ctx, nextAction)
}

// Creates a new Router with the given ContextFactory.
func New[T Action](actionFactory func(context.Context, *RouterContext[T])) *Router[T] {
	return &Router[T]{
		actionFactory: actionFactory,
		routes:        make([]*Route[T], 0),
	}
}

func (router *Router[T]) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var handler MiddlewareFunc

	handler = func(ctx context.Context, mctx *MiddlewareContext) {
		ok, params, routeHandler := router.dispatch(ctx, r)

		var mediumHandler func(ctx context.Context, a Action)

		if ok {
			mediumHandler = func(ctx context.Context, action Action) {
				routerContext := &RouterContext[T]{action: action, next: routeHandler}
				router.actionFactory(ctx, routerContext)
			}
		}

		if !ok {
			mediumHandler = func(ctx context.Context, action Action) {
				next := func(ctx context.Context, action T) {
					if router.missingRoute != nil {
						router.missingRoute(ctx, action)
					} else {
						action.ResponseWriter().WriteHeader(http.StatusNotFound)
						_, _ = action.Write([]byte("404 not found"))
					}
				}

				routerContext := &RouterContext[T]{action: action, next: next}
				router.actionFactory(ctx, routerContext)
			}
		}

		action := NewAction(rw, r, params)
		mediumHandler(ctx, action)
	}

	middlewares := &MiddlewareContext{
		Request:        r,
		ResponseWriter: rw,
		middlewares:    append(router.middleware, handler),
	}

	middlewares.Run()
}

func (router *Router[T]) dispatch(ctx context.Context, r *http.Request) (bool, map[string]string, func(context.Context, T)) {
	if route, params := router.routeFor(r); route != nil {

		return true, params, route.handler
		// return true, params, func(ctx context.Context, action T) {
		// 	router.actionFactory(ctx, action, func(ctx context.Context, action T) {
		// 		route.handler(ctx, action)
		// 	})
		// }
	}

	for _, group := range router.groups {
		if ok, params, handler := group.dispatch(ctx, r); ok {
			return true, params, handler
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
func (r *Router[T]) Use(middleware MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware)
}

func (r *Router[T]) register(group dispatchable[T]) {
	r.groups = append(r.groups, group)
}

func (r *Router[T]) prefix() string { return "" }

var _ registerable[Action] = (*Router[Action])(nil)
