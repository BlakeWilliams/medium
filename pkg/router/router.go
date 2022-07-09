package router

import (
	"context"
	"net/http"
)

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
type Middleware func(ctx context.Context, c Action, next MiddlewareFunc)

// A function that handles a request.
type HandlerFunc[C any] func(context.Context, C)

// Convenience type for middleware handlers
type MiddlewareFunc = HandlerFunc[Action]

// AroundHandler represents a function that wraps a given route handler. This is
// similar to middleware, but has access to the custom Action type and is called
// after the middleware layer.
type AroundHandler[T Action] func(ctx context.Context, action T, cb func(context.Context))

// ActionFactory is a function that returns a new context for each request.
// This is the entrypoint for the router and can be used to setup request data
// like fetching the current user, reading session data, etc.
type ActionFactory[T any] func(Action) T

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T Action] struct {
	routes         []*Route[T]
	middleware     []Middleware
	aroundHandlers []AroundHandler[T]
	actionFactory  ActionFactory[T]
	// Called when no route matches the request. Useful for rendering 404 pages.
	missingRoute HandlerFunc[T]
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

	matchingRoute, params := router.routeFor(r)
	var handler HandlerFunc[Action]

	// TODO - there's no reason we need to re-build the middleware stack and
	// around stack each request.
	if matchingRoute != nil {
		handler = func(ctx context.Context, baseAction Action) {
			action := router.actionFactory(baseAction)
			router.wrapHandler(matchingRoute.handler)(ctx, action)
		}
	} else {
		handler = func(ctx context.Context, baseAction Action) {
			action := router.actionFactory(baseAction)

			h := func(ctx context.Context, action T) {
				if router.missingRoute != nil {
					router.missingRoute(ctx, action)
				} else {
					action.Response().WriteHeader(http.StatusNotFound)
					_, _ = action.Write([]byte("404 not found"))
				}
			}

			router.wrapHandler(h)(ctx, action)
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

func (router *Router[T]) routeFor(r *http.Request) (*Route[T], map[string]string) {
	for _, route := range router.routes {
		if ok, params := route.IsMatch(r); ok {
			return route, params
		}
	}

	return nil, nil
}

func (r *Router[T]) wrapHandler(handler HandlerFunc[T]) HandlerFunc[T] {
	for i := len(r.aroundHandlers) - 1; i >= 0; i-- {
		newHandler := func(handler AroundHandler[T], next HandlerFunc[T]) func(ctx context.Context, ac T) {
			return func(ctx context.Context, ac T) {
				handler(ctx, ac, func(newctx context.Context) {
					next(newctx, ac)
				})
			}
		}(r.aroundHandlers[i], handler)

		handler = newHandler
	}

	return handler
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
