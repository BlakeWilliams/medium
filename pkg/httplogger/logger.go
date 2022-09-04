package httplogger

import (
	"time"

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/pkg/mlog"
)

// Middleware accepts an ErrorHandler and returns a medium.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(action medium.Action, next medium.MiddlewareFunc) {
	start := time.Now()

	mlog.Info(
		action.Context(),
		"Handling request",
		mlog.Fields{
			"method": action.Request().Method,
			"path":   action.Request().URL.Path,
		},
	)
	next(action)
	mlog.Info(
		action.Context(),
		"Request served",
		mlog.Fields{
			"method":   action.Request().Method,
			"path":     action.Request().URL.Path,
			"duration": time.Since(start).String(),
			"status":   action.Status(),
		},
	)
}

// Sets the given logger on context so it's available to future middleware
func ProviderMiddleware(logger mlog.Logger) medium.Middleware {
	return func(action medium.Action, next medium.MiddlewareFunc) {
		action.WithContext(mlog.Inject(action.Context(), logger))
		next(action)
	}
}
