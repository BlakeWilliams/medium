package controller

import (
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
	// that allows it to run code before the route handler is called via next.
	Before[T any] interface {
		Before(r *medium.Request[T], next medium.HandlerFunc[T]) medium.Response
	}
)

// Mount registers the routes from the controller with the provided router or
// routable (group, subrouter, etc). If the controller implements Before
// interface then the Before method will be called before the route handler.
func Mount[T any](r medium.Routable[T], controller Routable[T]) {
	routes := controller.RegisterRoutes()

	// TODO: Add validation that Before has the correct signature if it's
	// defined using reflection.
	if _, ok := any(controller).(Before[T]); ok {
		for key, handler := range routes {
			routes[key] = handler
			routes[key] = func(r *medium.Request[T]) medium.Response {
				return any(controller).(Before[T]).Before(r, handler)
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
