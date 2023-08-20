package medium

import (
	"context"
	"io"
	"net/http"
)

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
// type Middleware func(c Action, next HandlerFunc[Action])
type Middleware func(http.ResponseWriter, *http.Request, http.HandlerFunc)

// A function that handles a request.
type HandlerFunc[T any] func(context.Context, *Request[T]) Response

// BeforeFunc is a function that is called before the action is executed.
type BeforeFunc[T any] (func(ctx context.Context, req *Request[T], next Next) Response)

// Next is a function that calls the next BeforeFunc or HandlerFunc in the
// chain. It accepts a context and returns a Response.
type Next func(ctx context.Context) Response

// routeable is used to ensure parity between RouteGroup and Router
type routable[ParentData any, Data any] interface {
	Match(method string, path string, handler HandlerFunc[Data])
	Get(path string, handler HandlerFunc[Data])
	Post(path string, handler HandlerFunc[Data])
	Put(path string, handler HandlerFunc[Data])
	Patch(path string, handler HandlerFunc[Data])
	Delete(path string, handler HandlerFunc[Data])
	Before(before BeforeFunc[Data])
}

var _ routable[NoData, NoData] = (*RouteGroup[NoData, NoData])(nil)
var _ routable[NoData, NoData] = (*Router[NoData])(nil)

// Convenience type for middleware handlers
// type MiddlewareFunc = HandlerFunc[Action]

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T any] struct {
	middlewares  []Middleware
	routeGroup   *RouteGroup[NoData, T]
	missingRoute HandlerFunc[T]
}

// Creates a new Router with the given action creator used to create the application's root type.
func New[T any](dataCreator func(*RootRequest) T) *Router[T] {
	return &Router[T]{
		routeGroup: &RouteGroup[NoData, T]{
			routes: make([]*Route[T], 0),
			dataCreator: func(ctx context.Context, r *RootRequest) (context.Context, T) {
				return ctx, dataCreator(r)
			},
		},
	}
}

// NewWithContext behaves the same as New, but is passed a context and expects
// a context to be returned from the data creator.
func NewWithContext[T any](dataCreator func(context.Context, *RootRequest) (context.Context, T)) *Router[T] {
	return &Router[T]{
		routeGroup: &RouteGroup[NoData, T]{
			routes: make([]*Route[T], 0),
			dataCreator: func(ctx context.Context, r *RootRequest) (context.Context, T) {
				return dataCreator(ctx, r)
			},
		},
	}
}

func (router *Router[T]) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var handler http.HandlerFunc

	handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rootRequest := &RootRequest{originalRequest: r}
		ok, routeData, routeHandler := router.routeGroup.dispatch(rootRequest)

		ctx, data := router.routeGroup.dataCreator(r.Context(), rootRequest)
		newReq := NewRequest(rootRequest.originalRequest, data, routeData)

		var mediumHandler func(context.Context) Response

		if !ok {
			mediumHandler = func(ctx context.Context) Response {
				if router.missingRoute == nil {
					return StringResponse(http.StatusNotFound, "404 not found")
				}

				return router.missingRoute(
					ctx,
					newReq,
				)
			}
		} else {
			mediumHandler = func(ctx context.Context) Response {
				return routeHandler(ctx, rootRequest)
			}
		}

		res := mediumHandler(ctx)

		for key, values := range res.Header() {
			for _, value := range values {
				rw.Header().Add(key, value)
			}
		}
		if res := res.Status(); res != 0 {
			rw.WriteHeader(res)
		}
		if res.Body() != nil {
			io.Copy(rw, res.Body())
		}
	})

	for i := len(router.middlewares) - 1; i >= 0; i-- {
		middleware := router.middlewares[i]
		nextHandler := handler

		handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			middleware(rw, r, nextHandler)
		})
	}

	handler.ServeHTTP(rw, r)
}

// Match is used to add a new Route to the Router
func (r *Router[T]) Match(method string, path string, handler HandlerFunc[T]) {
	r.routeGroup.Match(method, path, handler)
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
	r.middlewares = append(r.middlewares, middleware)
}

var _ registerable[NoData] = (*Router[NoData])(nil)

func (r *Router[T]) register(group dispatchable[T]) {
	r.routeGroup.register(group)
}

func (r *Router[T]) prefix() string {
	return r.routeGroup.prefix()
}

func (r *Router[T]) Before(before BeforeFunc[T]) {
	r.routeGroup.Before(before)
}
