package httplogger

import (
	"net/http"
	"time"

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/mlog"
)

// Determines if status is available and loggable
type Statusable interface {
	Status() int
}

// Middleware accepts an ErrorHandler and returns a medium.Middleware that will
// rescue errors that happen in middlewares that are called after it.
func Middleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	mlog.Info(
		r.Context(),
		"Handling request",
		mlog.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		},
	)
	next(rw, r)
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
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		ctx := mlog.Inject(r.Context(), logger)
		next(rw, r.WithContext(ctx))
	}
}
