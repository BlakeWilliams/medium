package medium

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type groupAction struct {
	*BaseAction
}

func TestGroup(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Use(func(ctx context.Context, a Action, next MiddlewareFunc) {
		a.Response().Header().Add("x-from-middleware", "wow")
		next(ctx, a)
	})

	group := NewGroup(func(ctx context.Context, ba *BaseAction, next func(context.Context, *groupAction)) {
		action := &groupAction{BaseAction: ba}
		next(ctx, action)
	})
	group.Get("/hello/:name", func(ctx context.Context, c *groupAction) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	router.Register(group)

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func TestGroup_Context(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Use(func(ctx context.Context, a Action, next MiddlewareFunc) {
		ctx = context.WithValue(ctx, "foo", 1)
		a.Response().Header().Add("x-from-middleware", "wow")
		next(ctx, a)
	})

	group := NewGroup(func(ctx context.Context, ba *BaseAction, next func(context.Context, *groupAction)) {
		require.Equal(t, 1, ctx.Value("foo"))
		ctx = context.WithValue(ctx, "foo", 2)
		action := &groupAction{BaseAction: ba}
		next(ctx, action)
	})
	group.Get("/hello/:name", func(ctx context.Context, c *groupAction) {
		require.Equal(t, 2, ctx.Value("foo"))
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	router.Register(group)

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}
