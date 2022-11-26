package httpmethod

import (
	"net/http"
	"strings"

	"github.com/blakewilliams/medium"
)

// RewriteMiddleware rewrites the HTTP method based on the _method parameter
// passed when the request type is POST. This is useful when working with HTTP
// forms since form only supports GET and POST methods.
func RewriteMiddleware(action medium.Action, next medium.MiddlewareFunc) {
	if action.Request().Method == http.MethodPost {
		action.Request().ParseForm()

		formValue := strings.ToUpper(action.Request().Form.Get("_method"))
		action.Request().Method = formValue
	}

	next(action)
}
