package logger

import (
	"context"
	"time"

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/pkg/mlog"
)

// Middleware accepts an ErrorHandler and returns a medium.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(ctx context.Context, action medium.Action, next medium.MiddlewareFunc) {
	start := time.Now()

	mlog.Info(
		ctx,
		"Handling request",
		mlog.Fields{"path": action.Request().URL.Path},
	)
	next(ctx, action)
	mlog.Info(
		ctx,
		"Request served",
		mlog.Fields{
			"path":     action.Request().URL.Path,
			"duration": time.Since(start).String(),
			"status":   action.Status(),
		},
	)
}

// Sets the given logger on context so it's available to future middleware
func ProviderMiddleware(logger mlog.Logger) medium.Middleware {
	return func(ctx context.Context, action medium.Action, next medium.MiddlewareFunc) {
		ctx = mlog.Inject(ctx, logger)
		next(ctx, action)
	}
}
