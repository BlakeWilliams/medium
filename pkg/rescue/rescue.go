package rescue

import (
	"fmt"

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/pkg/mlog"
)

// Logger interface that is required by the middleware.
type Logger interface {
	Errorf(format string, v ...any)
	Fatalf(format string, v ...any)
}

//  An ErrorHandler is a function that is called when an error occurs.
type ErrorHandler func(medium.Action, error)

// Middleware accepts an ErrorHandler and returns a medium.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(handler ErrorHandler) medium.Middleware {
	return func(action medium.Action, next medium.MiddlewareFunc) {
		defer func() {
			err := recover()
			if err != nil {
				switch err.(type) {
				case error:
					mlog.Error(action.Context(), "rescued error in middleware", mlog.Fields{"error": err.(error)})
					handler(action, err.(error))
				default:
					mlog.Error(action.Context(), "rescued non-error in middleware", mlog.Fields{})
					handler(action, fmt.Errorf("Panic rescued: %v", err))
				}
			}
		}()

		next(action)
	}
}
