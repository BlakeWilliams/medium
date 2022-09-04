package medium

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Use(func(a Action, next MiddlewareFunc) {
		a.Response().Header().Add("x-from-middleware", "wow")
		next(a)
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

func TestRouter_Post(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Post("/hello/:name", func(c *BaseAction) {
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

	router.Use(func(a Action, next HandlerFunc[Action]) {
		a.Response().Header().Add("x-from-middleware", "wow")
		next(a)
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

func ExampleRouter_Register() {
	// create a router for the application
	router := New(DefaultActionFactory)

	// define a middleware that sets the x-powered-by header
	router.Use(func(a Action, next MiddlewareFunc) {
		a.Response().Header().Add("x-powered-by", "medium")
		next(a)
	})

	// define a new group
	group := NewGroup(func(ba *BaseAction, next func(*groupAction)) {
		action := &groupAction{BaseAction: ba}
		next(action)
	})
	// declare a route on the group
	group.Get("/hello/:name", func(c *groupAction) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	// register the group with the router
	router.Register(group)
}
