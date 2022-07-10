package medium

import (
	"context"
	"net/http"
)

type dispatchable[T Action] interface {
	dispatch(rw http.ResponseWriter, r *http.Request) (bool, map[string]string, func(context.Context, T))
}

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
type Middleware func(ctx context.Context, c Action, next MiddlewareFunc)

// A function that handles a request.
type HandlerFunc[C any] func(context.Context, C)

// Convenience type for middleware handlers
type MiddlewareFunc = HandlerFunc[Action]

// ActionFactory is a function that returns a new context for each request.
// This is the entrypoint for the router and can be used to setup request data
// like fetching the current user, reading session data, etc.
type ActionFactory[T any] func(context.Context, Action, func(context.Context, T))

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T Action] struct {
	routes        []*Route[T]
	middleware    []Middleware
	actionFactory ActionFactory[T]
	// Called when no route matches the request. Useful for rendering 404 pages.
	missingRoute HandlerFunc[T]

	groups []dispatchable[T]
}

// Creates a new Router with the given ContextFactory.
func New[T Action](actionFactory ActionFactory[T]) *Router[T] {
	return &Router[T]{
		actionFactory: actionFactory,
		routes:        make([]*Route[T], 0),
	}
}

func (router *Router[T]) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	ok, params, handler := router.dispatch(rw, r)

	if !ok {
		handler = func(ctx context.Context, action Action) {
			router.actionFactory(ctx, action, func(ctx context.Context, action T) {
				if router.missingRoute != nil {
					router.missingRoute(ctx, action)
				} else {
					action.Response().WriteHeader(http.StatusNotFound)
					_, _ = action.Write([]byte("404 not found"))
				}
			})
		}
	}

	action := NewAction(rw, r, params)
	next := handler

	for i := len(router.middleware) - 1; i >= 0; i-- {
		newNext := func(next HandlerFunc[Action], middleware Middleware) func(ctx context.Context, baseAction Action) {
			return func(ctx context.Context, baseAction Action) {
				middleware(ctx, baseAction, next)
			}
		}(next, router.middleware[i])
		next = newNext
	}

	next(ctx, action)
}

func (router *Router[T]) dispatch(rw http.ResponseWriter, r *http.Request) (bool, map[string]string, func(context.Context, Action)) {
	if route, params := router.routeFor(r); route != nil {
		return true, params, func(ctx context.Context, action Action) {
			router.actionFactory(ctx, action, func(ctx context.Context, action T) {
				route.handler(ctx, action)
			})
		}
	}

	for _, group := range router.groups {
		if ok, params, handler := group.dispatch(rw, r); ok {
			return true, params, func(ctx context.Context, action Action) {
				router.actionFactory(ctx, action, func(ctx context.Context, action T) {
					handler(ctx, action)
				})
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

// Registers a group of routes that will be routed to in addition to the routes
// defined on router.
//
// Use NewGroup to create a new group. Groups can be nested under routers or
// within other groups. This enables the creation of context specific actions
// that can inherit from their parent actions.
//
// Diagram of what the "inheritance" chain can look like:
//     router[GlobalAction]
// 	   |
// 	   |_Group[GlobalAction, LoggedInAction]
// 	   |
// 	   |-Group[LoggedInAction, Teams]
// 	   |
// 	   |-Group[LoggedInAction, Settings]
// 	   |
// 	   |-Group[LoggedInAction, Admin]
// 	     |
// 	     |_ Group[Admin, SuperAdmin]
func (r *Router[T]) Register(group dispatchable[T]) {
	r.groups = append(r.groups, group)
}

func (r *Router[T]) Base(path string) *RouteBase[T] {
	return &RouteBase[T]{
		router: r,
	}
}
