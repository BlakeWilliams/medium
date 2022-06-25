package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Use(func(c Action, next MiddlewareFunc) {
		c.Response().Header().Add("x-from-middleware", "wow")
		next(c)
	})

	router.Get("/hello/:name", func(c *BaseAction) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.Run(rw, req)

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

	router.Run(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
}

func TestRouter_MissingRoute_NoHandler(t *testing.T) {
	router := New(DefaultActionFactory)

	req := httptest.NewRequest(http.MethodGet, "/where/do/i/go", nil)
	rw := httptest.NewRecorder()

	router.Run(rw, req)

	require.Equal(t, "404 not found", rw.Body.String())
	require.Equal(t, 404, rw.Result().StatusCode)
}

func TestRouter_MissingRoute_WithHandler(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Missing(func(c *BaseAction) {
		c.Write([]byte("Sorry, can't find that page."))
	})

	req := httptest.NewRequest(http.MethodGet, "/where/do/i/go", nil)
	rw := httptest.NewRecorder()

	router.Run(rw, req)

	require.Equal(t, "Sorry, can't find that page.", rw.Body.String())
	require.Equal(t, 404, rw.Result().StatusCode)
}

type MyAction struct {
	Action
	Data int
}

func TestCustomActionType(t *testing.T) {
	router := New(func(a Action) *MyAction {
		return &MyAction{Action: a, Data: 1}
	})

	router.Use(func(c Action, next HandlerFunc[Action]) {
		c.Response().Header().Add("x-from-middleware", "wow")
		next(c)
	})

	router.Get("/hello/:name", func(c *MyAction) {
		c.Write([]byte(fmt.Sprintf("hello %s, data %d", c.Params()["name"], c.Data)))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.Run(rw, req)

	require.Equal(t, "hello Fox Mulder, data 1", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func TestCustomActionType_AroundHandler(t *testing.T) {
	router := New(func(a Action) *MyAction {
		return &MyAction{Action: a, Data: 1}
	})

	router.Use(func(c Action, next HandlerFunc[Action]) {
		c.Response().Header().Add("x-from-middleware", "wow")
		next(c)
	})

	router.Around(func(c *MyAction, next func()) {
		c.Data += 1
		next()
		c.Data += 5
	})

	router.Around(func(c *MyAction, next func()) {
		c.Data *= 2
		next()
		c.Data += 5
	})

	router.Get("/hello/:name", func(c *MyAction) {
		c.Write([]byte(fmt.Sprintf("hello %s, data %d", c.Params()["name"], c.Data)))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.Run(rw, req)

	// Data should be 4, since it starts at 1, first middleware adds 1, second
	// middleware multiplies by 2
	require.Equal(t, "hello Fox Mulder, data 4", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}
