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

	Before[T any] interface {
		Before(r *medium.Request[T], next medium.HandlerFunc[T]) medium.Response
	}
)

func Mount[T any](r medium.Routable[T], controller Routable[T]) {
	routes := controller.RegisterRoutes()

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
