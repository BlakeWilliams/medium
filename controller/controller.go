package controller

import (
	"reflect"
	"strings"

	"github.com/blakewilliams/medium"
)

type (
	// Routable is used to ensure that the controller can register routes which
	// will be registered with the router via medium.Routable
	Routable[T any] interface {
		RegisterRoutes() medium.Routes[T]
	}

	// Before is an optional interface that can be implemented by a controller
	// that allows it to run code before the route handler is called via nMext.
	Before[T any] interface {
		BeforeHandler(r *medium.Request[T], next medium.HandlerFunc[T]) medium.Response
	}
)

// Mount registers the routes from the controller with the provided router or
// routable (group, subrouter, etc). If the controller implements BeforeHandler
// interface then the BeforeHandler method will be called before the route
// handler.
func Mount[T any](r medium.Routable[T], controller Routable[T]) {
	routes := controller.RegisterRoutes()

	// check if type has BeforeHandler method, if it does validate it has the correct signature
	t := reflect.TypeOf(controller)
	if before, ok := t.MethodByName("BeforeHandler"); ok {
		if before.Type.NumIn() != 2 && before.Type.NumOut() != 1 {
			panic("Before method must have 2 arguments and 1 return value")
		}

		// In(0) is the receiver, so we want to check the first argument
		if before.Type.In(1) != reflect.TypeOf(&medium.Request[T]{}) {
			panic("BeforeHandler method first argument must be *medium.Request")
		}

		if before.Type.In(2) != reflect.TypeOf(medium.HandlerFunc[T](nil)) {
			panic("BeforeHandler method second argument must be medium.HandlerFunc")
		}
	}

	if _, ok := any(controller).(Before[T]); ok {
		for key, handler := range routes {
			routes[key] = handler
			routes[key] = func(r *medium.Request[T]) medium.Response {
				return any(controller).(Before[T]).BeforeHandler(r, handler)
			}
		}
	}

	for key, handler := range routes {
		parts := strings.Split(key, " ")
		r.Match(
			strings.TrimSpace(parts[0]),
			strings.TrimSpace(parts[1]),
			handler,
		)
	}
}
