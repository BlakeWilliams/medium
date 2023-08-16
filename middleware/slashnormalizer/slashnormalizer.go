// Package slashnormalizer provides a middleware that normalizes the URL path
// by removing duplicate slashes and redirects that normalized URL path.
package slashnormalizer

import (
	"net/http"
	"strings"
)

// Middleware normalizes the URL path by removing duplicate slashes and
// redirected to the normalized URL path.
func Middleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if strings.Contains(r.URL.Path, "//") {
		r.URL.Path = strings.Replace(r.URL.Path, "//", "/", -1)

		w.Header().Set("Location", r.URL.String())
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	next(w, r)
}
