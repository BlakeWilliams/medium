package medium

import (
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

	router.Get("/hello/:name", func(c *BaseAction) {
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
	router := New(DefaultActionCreator)

	for name, tc := range testCases {
		path := reflect.ValueOf("/")
		var handler HandlerFunc[*BaseAction] = func(c *BaseAction) {
			c.ResponseWriter().WriteHeader(http.StatusOK)
			c.Write([]byte("hello"))
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

	router.Post("/hello/:name", func(c *BaseAction) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
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

	router.Missing(func(c *BaseAction) {
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
	router := New(func(a Action, next func(*MyAction)) {
		action := &MyAction{Action: a, Data: 1}

		next(action)
	})

	router.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Add("x-from-middleware", "wow")
		next(rw, r)
	})

	router.Get("/hello/:name", func(c *MyAction) {
		c.Write([]byte(fmt.Sprintf("hello %s, data %d", c.Params()["name"], c.Data)))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder, data 1", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}
