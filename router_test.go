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

func TestHappyPath(t *testing.T) {
	router := New(DefaultActionCreator)

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	router.Get("/hello/:name", func(c Request[NoData]) {
		c.Response().Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func TestGroup_RouteMethods(t *testing.T) {
	testCases := map[string]struct {
		method string
	}{
		"Get":    {method: http.MethodGet},
		"Post":   {method: http.MethodPost},
		"Put":    {method: http.MethodPut},
		"Patch":  {method: http.MethodPatch},
		"Delete": {method: http.MethodDelete},
	}
	router := New(DefaultActionCreator)

	for name, tc := range testCases {
		path := reflect.ValueOf("/")
		var handler HandlerFunc[NoData] = func(r Request[NoData]) {
			r.Response().WriteHeader(http.StatusOK)
			r.Response().Write([]byte("hello"))
		}

		handlerValue := reflect.ValueOf(handler)

		t.Run(name, func(t *testing.T) {
			reflect.ValueOf(router).MethodByName(name).Call([]reflect.Value{path, handlerValue})

			req := httptest.NewRequest(tc.method, "/", nil)
			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			require.Equal(t, 200, rw.Code)
		})
	}
}

func TestRouter_Post(t *testing.T) {
	router := New(DefaultActionCreator)

	router.Post("/hello/:name", func(c Request[NoData]) {
		c.Response().Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	req := httptest.NewRequest(http.MethodPost, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
}

func TestRouter_MissingRoute_NoHandler(t *testing.T) {
	router := New(DefaultActionCreator)

	req := httptest.NewRequest(http.MethodGet, "/where/do/i/go", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "404 not found", rw.Body.String())
	require.Equal(t, 404, rw.Result().StatusCode)
}

func TestRouter_MissingRoute_WithHandler(t *testing.T) {
	router := New(DefaultActionCreator)

	router.Missing(func(c Request[NoData]) {
		c.Response().WriteHeader(http.StatusNotFound)
		c.Response().Write([]byte("Sorry, can't find that page."))
	})

	req := httptest.NewRequest(http.MethodGet, "/where/do/i/go", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "Sorry, can't find that page.", rw.Body.String())
	require.Equal(t, 404, rw.Result().StatusCode)
}

func TestCustomActionType(t *testing.T) {
	router := New[*MyData](func(rootRequest RootRequest, next func(*MyData)) {
		next(&MyData{Value: 1})
	})

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	router.Get("/hello/:name", func(c Request[*MyData]) {
		c.Response().Write([]byte(fmt.Sprintf("hello %s, data %d", c.Params()["name"], c.Data.Value)))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder, data 1", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

type myResponseWriter struct {
	orw http.ResponseWriter
}

var _ http.ResponseWriter = (*myResponseWriter)(nil)

func (mrw *myResponseWriter) Header() http.Header         { return mrw.orw.Header() }
func (mrw *myResponseWriter) WriteHeader(s int)           { mrw.orw.WriteHeader(s) }
func (mrw *myResponseWriter) Write(b []byte) (int, error) { return mrw.orw.Write(b) }

func TestCustomResponseWriter(t *testing.T) {
	router := New(DefaultActionCreator)
	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		mrw := &myResponseWriter{orw: rw}
		next(mrw, r)
	})

	called := false
	router.Get("/", func(ba Request[NoData]) {
		require.IsType(t, &myResponseWriter{}, ba.Response().Writer())
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	require.True(t, called)
}

func TestCustomResponse(t *testing.T) {
	router := New(DefaultActionCreator)
	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		ctx := context.WithValue(r.Context(), "foo", "bar")
		next(rw, r.WithContext(ctx))
	})

	called := false
	router.Get("/", func(ba Request[NoData]) {
		require.Equal(t, "bar", ba.Request().Context().Value("foo"))
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()
	router.ServeHTTP(rw, req)

	require.True(t, called)
}
