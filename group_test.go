package medium

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type MyData struct {
	Value int
}

func TestGroup_Routes(t *testing.T) {
	testCases := map[string]struct {
		method string
	}{
		"Get":    {method: http.MethodGet},
		"Post":   {method: http.MethodPost},
		"Put":    {method: http.MethodPut},
		"Patch":  {method: http.MethodPatch},
		"Delete": {method: http.MethodDelete},
	}
	router := New(WithNoData)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r)
	})

	group := NewGroup(router, func(og NoData) MyData {
		return MyData{Value: 1}
	})

	for name, tc := range testCases {
		path := reflect.ValueOf("/")
		handler := reflect.ValueOf(func(ctx context.Context, r Request[MyData]) Response {
			return StringResponse(http.StatusOK, "hello")
		})

		t.Run(name, func(t *testing.T) {
			reflect.ValueOf(group).MethodByName(name).Call([]reflect.Value{path, handler})

			req := httptest.NewRequest(tc.method, "/", nil)
			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			require.Equal(t, 200, rw.Code)
		})
	}
}

func TestGroup(t *testing.T) {
	router := New(WithNoData)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	group := NewGroup(router, func(_ NoData) MyData {
		return MyData{Value: 1}
	})
	group.Get("/hello/:name", func(ctx context.Context, r Request[MyData]) Response {
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func TestGroup_Subgroup(t *testing.T) {
	router := New(WithNoData)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	group := NewGroup[NoData, MyData, registerable[NoData]](router, func(data NoData) MyData {
		return MyData{Value: 1}
	})

	subgroup := NewGroup(group, func(data MyData) MyData {
		data.Value += 1
		return data
	})

	subgroup.Get("/hello/:name", func(ctx context.Context, c Request[MyData]) Response {
		require.Equal(t, 2, c.Data.Value)
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", c.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func TestGroup_Subrouter(t *testing.T) {
	router := New(WithNoData)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	group := Subrouter(router, "/foo", func(_ NoData) MyData {
		return MyData{Value: 1}
	})

	subgroup := Subrouter(group, "/bar", func(data MyData) MyData {
		data.Value += 1
		return data
	})

	subgroup.Get("/hello/:name", func(ctx context.Context, c Request[MyData]) Response {
		require.Equal(t, 2, c.Data.Value)
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", c.Params()["name"]))
	})

	subsubgroup := Subrouter(subgroup, "/baz", func(data MyData) MyData {
		data.Value += 1
		return data
	})

	subsubgroup.Get("/hello/:name", func(ctx context.Context, c Request[MyData]) Response {
		require.Equal(t, 3, c.Data.Value)
		return StringResponse(http.StatusOK, fmt.Sprintf("hello again %s", c.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/foo/bar/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))

	req = httptest.NewRequest(http.MethodGet, "/foo/bar/baz/hello/Fox%20Mulder", nil)
	rw = httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello again Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}
