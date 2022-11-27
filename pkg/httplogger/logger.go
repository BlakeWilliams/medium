package httplogger

import (
	"context"
	"net/http"
	"time"

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/pkg/mlog"
)

// Determines if status is available and loggable
type Statusable interface {
	Status() int
}

// Middleware accepts an ErrorHandler and returns a medium.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(ctx context.Context, r *http.Request, rw http.ResponseWriter, next medium.NextMiddleware) {
	start := time.Now()

	mlog.Info(
		ctx,
		"Handling request",
		mlog.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		},
	)
	next(ctx, r, rw)
	fields := mlog.Fields{
		"method":   r.Method,
		"path":     r.URL.Path,
		"duration": time.Since(start).String(),
	}

	if status, ok := rw.(Statusable); ok {
		fields["status"] = status.Status()
	}

	mlog.Info(
		r.Context(),
		"Request served",
		fields,
	)
}

// Sets the given logger on context so it's available to future middleware
func ProviderMiddleware(logger mlog.Logger) medium.Middleware {
	return func(ctx context.Context, r *http.Request, rw http.ResponseWriter, next medium.NextMiddleware) {
		ctx = mlog.Inject(ctx, logger)
		next(ctx, r, rw)
	}
}
