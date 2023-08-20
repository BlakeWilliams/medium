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

	group := Group(router, func(og NoData) MyData {
		return MyData{Value: 1}
	})

	for name, tc := range testCases {
		path := reflect.ValueOf("/")
		handler := reflect.ValueOf(func(ctx context.Context, r *Request[MyData]) Response {
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

	group := Group(router, func(_ NoData) MyData {
		return MyData{Value: 1}
	})
	group.Get("/hello/:name", func(ctx context.Context, r *Request[MyData]) Response {
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

	group := Group[NoData, MyData, registerable[NoData]](router, func(data NoData) MyData {
		return MyData{Value: 1}
	})

	subgroup := Group(group, func(data MyData) MyData {
		data.Value += 1
		return data
	})

	subgroup.Get("/hello/:name", func(ctx context.Context, r *Request[MyData]) Response {
		require.Equal(t, 2, r.Data.Value)
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
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

	group := SubRouter(router, "/foo", func(_ NoData) MyData {
		return MyData{Value: 1}
	})

	subgroup := SubRouter(group, "/bar", func(data MyData) MyData {
		data.Value += 1
		return data
	})

	subgroup.Get("/hello/:name", func(ctx context.Context, r *Request[MyData]) Response {
		require.Equal(t, 2, r.Data.Value)
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	subsubgroup := SubRouter(subgroup, "/baz", func(data MyData) MyData {
		data.Value += 1
		return data
	})

	subsubgroup.Get("/hello/:name", func(ctx context.Context, r *Request[MyData]) Response {
		require.Equal(t, 3, r.Data.Value)
		return StringResponse(http.StatusOK, fmt.Sprintf("hello again %s", r.Params()["name"]))
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

func TestGroup_Before(t *testing.T) {
	router := New(WithNoData)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	group := Group(router, func(_ NoData) MyData {
		return MyData{Value: 1}
	})

	called := false
	group.Before(func(ctx context.Context, r *Request[MyData], next Next) Response {
		called = true

		require.Equal(t, 1, r.Data.Value)
		r.Data.Value += 1
		return next(ctx)
	})

	group.Get("/hello/:name", func(ctx context.Context, r *Request[MyData]) Response {
		require.Equal(t, 2, r.Data.Value)
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.True(t, called)
	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func TestGroup_Before_NestedGroup(t *testing.T) {
	router := New(WithNoData)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	group := Group(router, func(_ NoData) MyData {
		return MyData{Value: 1}
	})

	called := false
	group.Before(func(ctx context.Context, r *Request[MyData], next Next) Response {
		called = true

		require.Equal(t, 1, r.Data.Value)
		r.Data.Value += 1
		return next(ctx)
	})

	subgroup := Group(group, func(data MyData) MyData {
		require.Equal(t, 2, data.Value)
		data.Value += 1
		return data
	})
	subgroup.Before(func(ctx context.Context, r *Request[MyData], next Next) Response {
		require.Equal(t, 3, r.Data.Value)
		r.Data.Value += 1
		return next(ctx)
	})

	subgroup.Get("/hello/:name", func(ctx context.Context, r *Request[MyData]) Response {
		require.Equal(t, 4, r.Data.Value)
		return StringResponse(http.StatusOK, fmt.Sprintf("hello %s", r.Params()["name"]))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.True(t, called)
	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}
