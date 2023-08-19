package medium

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteMatching(t *testing.T) {
	tests := map[string]struct {
		reqMethod   string
		reqPath     string
		routeMethod string
		routePath   string
		want        bool
		params      map[string]string
	}{
		"length check": {
			reqMethod:   "GET",
			reqPath:     "/foo",
			routeMethod: "GET",
			routePath:   "/foo/bar",
			want:        false,
			params:      nil,
		},
		"path mismatch": {
			reqMethod:   "GET",
			reqPath:     "/foo",
			routeMethod: "GET",
			routePath:   "/bar",
			want:        false,
			params:      nil,
		},
		"method mismatch": {
			reqMethod:   "CONNECT",
			reqPath:     "/foo",
			routeMethod: "GET",
			routePath:   "/foo",
			want:        false,
			params:      nil,
		},
		"valid root": {
			reqMethod:   "GET",
			reqPath:     "/",
			routeMethod: "GET",
			routePath:   "/",
			want:        true,
			params:      map[string]string{},
		},
		"valid basic route": {
			reqMethod:   "GET",
			reqPath:     "/foo",
			routeMethod: "GET",
			routePath:   "/foo",
			want:        true,
			params:      map[string]string{},
		},
		"valid long route": {
			reqMethod:   "GET",
			reqPath:     "/foo/baz/bar",
			routeMethod: "GET",
			routePath:   "/foo/baz/bar",
			want:        true,
			params:      map[string]string{},
		},
		"valid route params": {
			reqMethod:   "GET",
			reqPath:     "/foo/baz/bar?name=true",
			routeMethod: "GET",
			routePath:   "/foo/baz/bar",
			want:        true,
			params:      map[string]string{},
		},
		"valid dynamic route": {
			reqMethod:   "GET",
			reqPath:     "/hello/greg",
			routeMethod: "GET",
			routePath:   "/hello/:name",
			want:        true,
			params:      map[string]string{"name": "greg"},
		},
		"valid multi dynamic route": {
			reqMethod:   "GET",
			reqPath:     "/hello/greg/boston",
			routeMethod: "GET",
			routePath:   "/hello/:name/:location",
			want:        true,
			params:      map[string]string{"name": "greg", "location": "boston"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(tc.reqMethod, tc.reqPath, nil)
			root := RootRequest{originalRequest: req, response: response{responseWriter: httptest.NewRecorder()}}
			route := newRoute(tc.routeMethod, tc.routePath, func(Request[NoData]) Response {
				return OK()
			})

			got, params := route.IsMatch(root)

			assert.Equal(t, got, tc.want, "expected route to match")
			assert.Equal(t, params, tc.params, "expected route to match")
		})
	}
}
