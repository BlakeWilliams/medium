package httpmethod

import (
	"context"
	"net/http"
	"strings"

	"github.com/blakewilliams/medium"
)

// RewriteMiddleware rewrites the HTTP method based on the _method parameter
// passed when the request type is POST. This is useful when working with HTTP
// forms since form only supports GET and POST methods.
func RewriteMiddleware(ctx context.Context, r *http.Request, rw http.ResponseWriter, next medium.NextMiddleware) {
	if r.Method == http.MethodPost {
		if method := r.FormValue("_method"); method != "" {
			r.Method = strings.ToUpper(method)
		}
	}

	next(ctx, r, rw)
}
