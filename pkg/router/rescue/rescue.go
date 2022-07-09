package rescue

import (
	"context"
	"fmt"

	"github.com/blakewilliams/medium/pkg/mlog"
	"github.com/blakewilliams/medium/pkg/router"
)

// Logger interface that is required by the middleware.
type Logger interface {
	Errorf(format string, v ...any)
	Fatalf(format string, v ...any)
}

//  An ErrorHandler is a function that is called when an error occurs.
type ErrorHandler func(router.Action, error)

// Middleware accepts an ErrorHandler and returns a router.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(handler ErrorHandler) router.Middleware {
	return func(ctx context.Context, action router.Action, next router.MiddlewareFunc) {
		defer func() {
			err := recover()
			if err != nil {
				switch err.(type) {
				case error:
					mlog.Error(ctx, "rescued error in middleware", mlog.Fields{"error": err.(error)})
					handler(action, err.(error))
				default:
					mlog.Error(ctx, "rescued non-error in middleware", mlog.Fields{})
					handler(action, fmt.Errorf("Panic rescued: %v", err))
				}
			}
		}()

		next(ctx, action)
	}
}
