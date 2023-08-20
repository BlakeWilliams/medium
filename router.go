package medium

import (
	"context"
	"io"
	"net/http"
)

type dispatchable[T any] interface {
	dispatch(r RootRequest) (bool, *RouteData, func(context.Context, *Request[T]) Response)
}

var _ dispatchable[NoData] = (*Router[NoData])(nil)
var _ dispatchable[NoData] = (*RouteGroup[NoData, NoData])(nil)

// Middleware is a function that is called before the action is executed.
// See Router.Use for more information.
// type Middleware func(c Action, next HandlerFunc[Action])
type Middleware func(http.ResponseWriter, *http.Request, http.HandlerFunc)

// A function that handles a request.
type HandlerFunc[T any] func(context.Context, *Request[T]) Response

// A function that calls the next BeforeFunc or HandlerFunc in the chain.
type Next func(ctx context.Context) Response

// A function that is called before the action is executed.
type BeforeFunc[T any] (func(ctx context.Context, req *Request[T], next Next) Response)

// Convenience type for middleware handlers
// type MiddlewareFunc = HandlerFunc[Action]

// Router is a collection of Routes and is used to dispatch requests to the
// correct Route handler.
type Router[T any] struct {
	routes      []*Route[T]
	middlewares []Middleware
	befores     []BeforeFunc[T]
	dataCreator func(RootRequest) T
	// Called when no route matches the request. Useful for rendering 404 pages.
	missingRoute HandlerFunc[T]

	groups []dispatchable[T]
}

// Creates a new Router with the given action creator used to create the application's root type.
func New[T any](dataCreator func(RootRequest) T) *Router[T] {
	return &Router[T]{
		dataCreator: dataCreator,
		routes:      make([]*Route[T], 0),
	}
}

func (router *Router[T]) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var handler http.HandlerFunc

	handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rootRequest := RootRequest{originalRequest: r}
		ok, routeData, routeHandler := router.dispatch(rootRequest)

		data := router.dataCreator(rootRequest)
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
				return routeHandler(
					ctx,
					newReq,
				)
			}

			// Run before actions
			for i := len(router.befores) - 1; i >= 0; i-- {
				before := router.befores[i]
				nextHandler := mediumHandler

				mediumHandler = func(ctx context.Context) Response {

					return before(
						ctx,
						newReq,
						nextHandler,
					)
				}
			}
		}

		res := mediumHandler(r.Context())

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

func (router *Router[T]) dispatch(r RootRequest) (bool, *RouteData, func(context.Context, *Request[T]) Response) {
	if route, routeData := router.routeFor(r); route != nil {
		return true, routeData, route.handler
	}

	for _, group := range router.groups {
		if ok, routeData, handler := group.dispatch(r); ok {
			return true, routeData, handler
		}
	}

	return false, nil, nil
}

func (router *Router[T]) routeFor(r RootRequest) (*Route[T], *RouteData) {
	for _, route := range router.routes {
		if ok, params := route.IsMatch(r); ok {
			routeData := &RouteData{Params: params, HandlerPath: route.Raw}

			return route, routeData
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
	r.middlewares = append(r.middlewares, middleware)
}

var _ registerable[NoData] = (*Router[NoData])(nil)

func (r *Router[T]) register(group dispatchable[T]) {
	r.groups = append(r.groups, group)
}

func (r *Router[T]) prefix() string { return "" }

func (r *Router[T]) Before(before BeforeFunc[T]) {
	r.befores = append(r.befores, before)
}
