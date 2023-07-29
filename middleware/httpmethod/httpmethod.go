package httpmethod

import (
	"net/http"
	"strings"
)

// RewriteMiddleware rewrites the HTTP method based on the _method parameter
// passed when the request type is POST. This is useful when working with HTTP
// forms since form only supports GET and POST methods.
func RewriteMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method == http.MethodPost {
		if method := r.FormValue("_method"); method != "" {
			r.Method = strings.ToUpper(method)
		}
	}

	next(rw, r)
}
