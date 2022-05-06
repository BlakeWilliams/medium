package router

import (
	"net/http"

	"github.com/blakewilliams/medium/pkg/hooks"
)

type ServeEvent[T Controller] struct {
	Route Route[T]
	Controller T
}

type ControllerFactory[T Controller] func(http.ResponseWriter, *http.Request, map[string]string) T

type Router[T Controller] struct {
	Routes            []*Route[T]
	controllerCreator func(http.ResponseWriter, *http.Request, map[string]string) T
	serveHook *hooks.Emitter[ServeEvent[T]]
}

func New[T Controller](controllerFactory ControllerFactory[T]) *Router[T] {
	return &Router[T]{
		Routes:            make([]*Route[T], 0),
		controllerCreator: controllerFactory,
		serveHook: hooks.NewEmitter[ServeEvent[T]](),
	}
}

func (router *Router[T]) Run(rw http.ResponseWriter, r *http.Request) {
	for _, route := range router.Routes {
		if ok, params := route.IsMatch(r); ok {
			controller := router.controllerCreator(rw, r, params)

			payload := ServeEvent[T]{Route: *route, Controller: controller}
			router.serveHook.Emit(payload, func() {
				route.handler(controller)
			})
			break
		}
	}
}

func (r *Router[Controller]) Match(method string, path string, controller func(Controller)) {
	r.Routes = append(r.Routes, newRoute(method, path, controller))
}

func (r *Router[Controller]) Get(path string, controller func(Controller)) {
	r.Routes = append(r.Routes, newRoute(http.MethodGet, path, controller))
}

func (r *Router[T]) OnServe(handler func(hooks.Event[ServeEvent[T]])) {
	r.serveHook.Subscribe(handler)
}
