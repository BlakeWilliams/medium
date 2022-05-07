package router

import (
	"net/http"
)

type Middleware[T any] func(T, HandlerFunc[T])
type ContextFactory[T any] func(http.ResponseWriter, *http.Request, map[string]string) T

type Router[T Action] struct {
	Application T
	Routes      []*Route[T]
	middleware  []Middleware[T]
	contextFactory ContextFactory[T]
}

func New[T Action](contextFactory ContextFactory[T]) *Router[T] {
	return &Router[T]{
		contextFactory: contextFactory,
		Routes:      make([]*Route[T], 0),
	}
}

func (router *Router[T]) Run(rw http.ResponseWriter, r *http.Request) {
	var matchingRoute *Route[T]
	params := map[string]string{}

	for _, route := range router.Routes {
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

func (r *Router[T]) Match(method string, path string, handler HandlerFunc[T]) {
	r.Routes = append(r.Routes, newRoute(method, path, handler))
}

func (r *Router[T]) Get(path string, handler HandlerFunc[T]) {
	r.Routes = append(r.Routes, newRoute(http.MethodGet, path, handler))
}

func (r *Router[T]) Use(middleware func(c T, next HandlerFunc[T])) {
	r.middleware = append(r.middleware, middleware)
}
