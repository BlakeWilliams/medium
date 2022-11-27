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
	router := New(DefaultActionFactory)

	router.Use(func(ctx context.Context, r *http.Request, rw http.ResponseWriter, next NextMiddleware) {
		rw.Header().Add("x-from-middleware", "wow")
		next(ctx, r, rw)
	})

	router.Get("/hello/:name", func(ctx context.Context, c *BaseAction) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
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
	router := New(DefaultActionFactory)

	for name, tc := range testCases {
		path := reflect.ValueOf("/")
		handler := reflect.ValueOf(func(ctx context.Context, c *BaseAction) {
			c.ResponseWriter().WriteHeader(http.StatusOK)
			c.Write([]byte("hello"))
		})

		t.Run(name, func(t *testing.T) {
			reflect.ValueOf(router).MethodByName(name).Call([]reflect.Value{path, handler})

			req := httptest.NewRequest(tc.method, "/", nil)
			rw := httptest.NewRecorder()
			router.ServeHTTP(rw, req)

			require.Equal(t, 200, rw.Code)
		})
	}
}

func TestRouter_Post(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Post("/hello/:name", func(ctx context.Context, c *BaseAction) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	req := httptest.NewRequest(http.MethodPost, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
}

func TestRouter_MissingRoute_NoHandler(t *testing.T) {
	router := New(DefaultActionFactory)

	req := httptest.NewRequest(http.MethodGet, "/where/do/i/go", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "404 not found", rw.Body.String())
	require.Equal(t, 404, rw.Result().StatusCode)
}

func TestRouter_MissingRoute_WithHandler(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Missing(func(ctx context.Context, c *BaseAction) {
		c.ResponseWriter().WriteHeader(http.StatusNotFound)
		c.Write([]byte("Sorry, can't find that page."))
	})

	req := httptest.NewRequest(http.MethodGet, "/where/do/i/go", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "Sorry, can't find that page.", rw.Body.String())
	require.Equal(t, 404, rw.Result().StatusCode)
}

type MyAction struct {
	Action
	Data int
}

func TestCustomActionType(t *testing.T) {
	router := New(func(ctx context.Context, m Migrator[Action, *MyAction]) {
		action := &MyAction{Action: m.Action(), Data: 1}

		m.Next(ctx, action)
	})

	router.Use(func(ctx context.Context, r *http.Request, rw http.ResponseWriter, next NextMiddleware) {
		rw.Header().Add("x-from-middleware", "wow")
		next(ctx, r, rw)
	})

	router.Get("/hello/:name", func(ctx context.Context, c *MyAction) {
		c.Write([]byte(fmt.Sprintf("hello %s, data %d", c.Params()["name"], c.Data)))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder, data 1", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}
