package rescue

import (
	"fmt"

	"github.com/blakewilliams/medium/pkg/router"
)

// Logger interface that is required by the middleware.
type Logger interface {
	Errorf(format string, v ...any)
	Fatalf(format string, v ...any)
}

//  An ErrorHandler is a function that is called when an error occurs.
type ErrorHandler[T router.Action] func(T, error)

// Middleware accepts an ErrorHandler and returns a router.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware[T router.Action](handler ErrorHandler[T]) router.Middleware[T] {
	return func(c T, next router.HandlerFunc[T]) {
		defer func() {
			err := recover()
			if err != nil {
				switch err.(type) {
				case error:
					handler(c, err.(error))
				default:
					handler(c, fmt.Errorf("Panic rescued: %v", err))
				}
			}
		}()

		next(c)
	}
}
