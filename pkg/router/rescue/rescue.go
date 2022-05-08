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
type ErrorHandler func(*router.BaseAction, error)

// Middleware accepts an ErrorHandler and returns a router.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(handler ErrorHandler) router.Middleware {
	return func(c *router.BaseAction, next router.HandlerFunc[*router.BaseAction]) {
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
