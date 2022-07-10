package medium

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Use(func(ctx context.Context, a Action, next MiddlewareFunc) {
		a.Response().Header().Add("x-from-middleware", "wow")
		next(ctx, a)
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
	router := New(func(ctx context.Context, a Action, next func(context.Context, *MyAction)) {
		action := &MyAction{Action: a, Data: 1}

		next(ctx, action)
	})

	router.Use(func(ctx context.Context, a Action, next HandlerFunc[Action]) {
		a.Response().Header().Add("x-from-middleware", "wow")
		next(ctx, a)
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

func TestContextPropagation(t *testing.T) {
	router := New(func(ctx context.Context, a Action, next func(context.Context, *MyAction)) {
		action := &MyAction{Action: a, Data: 1}
		next(ctx, action)
	})

	router.Use(func(ctx context.Context, a Action, next HandlerFunc[Action]) {
		ctx = context.WithValue(ctx, "foo", "bar")
		a.Response().Header().Add("x-from-middleware", "wow")
		next(ctx, a)
	})

	router.Use(func(ctx context.Context, a Action, next HandlerFunc[Action]) {
		ctx = context.WithValue(ctx, "bar", "baz")
		// c.Data += 1
		next(ctx, a)
		// c.Data += 5
	})

	router.Use(func(ctx context.Context, a Action, next HandlerFunc[Action]) {
		ctx = context.WithValue(ctx, "baz", "qux")
		// c.Data *= 2
		next(ctx, a)
		// c.Data += 5
	})

	router.Get("/hello/:name", func(ctx context.Context, c *MyAction) {
		require.Equal(t, "bar", ctx.Value("foo"))
		require.Equal(t, "baz", ctx.Value("bar"))
		require.Equal(t, "qux", ctx.Value("baz"))
		c.Write([]byte(fmt.Sprintf("hello %s, data %d", c.Params()["name"], c.Data)))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	// Data should be 4, since it starts at 1, first middleware adds 1, second
	// middleware multiplies by 2
	// require.Equal(t, "hello Fox Mulder, data 4", rw.Body.String())
	require.Equal(t, "hello Fox Mulder, data 1", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func ExampleRouter_Register() {
	// create a router for the application
	router := New(DefaultActionFactory)

	// define a middleware that sets the x-powered-by header
	router.Use(func(ctx context.Context, a Action, next MiddlewareFunc) {
		a.Response().Header().Add("x-powered-by", "medium")
		next(ctx, a)
	})

	// define a new group
	group := NewGroup(func(ctx context.Context, ba *BaseAction, next func(context.Context, *groupAction)) {
		action := &groupAction{BaseAction: ba}
		next(ctx, action)
	})
	// declare a route on the group
	group.Get("/hello/:name", func(ctx context.Context, c *groupAction) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	// register the group with the router
	router.Register(group)
}
