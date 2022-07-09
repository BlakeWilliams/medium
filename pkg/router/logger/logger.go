package logger

import (
	"context"
	"time"

	"github.com/blakewilliams/medium/pkg/mlog"
	"github.com/blakewilliams/medium/pkg/router"
)

// Middleware accepts an ErrorHandler and returns a router.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(ctx context.Context, action router.Action, next router.MiddlewareFunc) {
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
func ProviderMiddleware(logger mlog.Logger) router.Middleware {
	return func(ctx context.Context, action router.Action, next router.MiddlewareFunc) {
		ctx = mlog.Inject(ctx, logger)
		next(ctx, action)
	}
}
