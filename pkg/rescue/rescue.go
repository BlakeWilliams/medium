package rescue

import (
	"context"
	"fmt"
	"net/http"

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/pkg/mlog"
)

// Logger interface that is required by the middleware.
type Logger interface {
	Errorf(format string, v ...any)
	Fatalf(format string, v ...any)
}

// An ErrorHandler is a function that is called when an error occurs.
type ErrorHandler func(context.Context, *http.Request, http.ResponseWriter, error)

// Middleware accepts an ErrorHandler and returns a medium.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(handler ErrorHandler) medium.MiddlewareFunc {
	return func(ctx context.Context, r *http.Request, rw http.ResponseWriter, next medium.NextMiddleware) {
		defer func() {
			err := recover()
			if err != nil {
				switch err.(type) {
				case error:
					mlog.Error(ctx, "rescued error in middleware", mlog.Fields{"error": err.(error)})
					handler(ctx, r, rw, err.(error))
				default:
					mlog.Error(ctx, "rescued non-error in middleware", mlog.Fields{"err": fmt.Sprintf("%v", err)})
					handler(ctx, r, rw, fmt.Errorf("Panic rescued: %v", err))
				}
			}
		}()

		next(ctx, r, rw)
	}
}
