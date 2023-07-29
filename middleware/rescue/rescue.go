package rescue

import (
	"fmt"
	"net/http"

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/mlog"
)

// Logger interface that is required by the middleware.
type Logger interface {
	Errorf(format string, v ...any)
	Fatalf(format string, v ...any)
}

// An ErrorHandler is a function that is called when an error occurs.
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// Middleware accepts an ErrorHandler and returns a medium.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(handler ErrorHandler) medium.Middleware {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		defer func() {
			rec := recover()
			if rec != nil {
				switch err := rec.(type) {
				case error:
					mlog.Error(r.Context(), "rescued error in middleware", mlog.Fields{"error": err})
					handler(rw, r, err)
				default:
					mlog.Error(r.Context(), "rescued non-error in middleware", mlog.Fields{"err": fmt.Sprintf("%v", err)})
					handler(rw, r, fmt.Errorf("Panic rescued: %v", err))
				}
			}
		}()

		next(rw, r)
	}
}
