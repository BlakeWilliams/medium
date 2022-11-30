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

type groupAction struct {
	value int
	*BaseAction
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
	router := New(DefaultActionFactory)

	router.Use(func(ctx context.Context, mctx *MiddlewareContext) {
		mctx.Next(ctx)
	})

	group := NewGroup(router, func(ctx context.Context, gctx *GroupContext[*BaseAction, *groupAction]) {
		action := &groupAction{BaseAction: gctx.Action()}
		gctx.Next(ctx, action)
	})

	for name, tc := range testCases {
		path := reflect.ValueOf("/")
		handler := reflect.ValueOf(func(ctx context.Context, c *groupAction) {
			c.ResponseWriter().WriteHeader(http.StatusOK)
			c.Write([]byte("hello"))
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
	router := New(DefaultActionFactory)

	router.Use(func(ctx context.Context, mctx *MiddlewareContext) {
		mctx.ResponseWriter.Header().Add("x-from-middleware", "wow")
		mctx.Next(ctx)
	})

	group := NewGroup(router, func(ctx context.Context, ba *BaseAction, next func(context.Context, *groupAction)) {
		action := &groupAction{BaseAction: ba}
		next(ctx, action)
	})
	group.Get("/hello/:name", func(ctx context.Context, c *groupAction) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func TestGroup_Subgroup(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Use(func(ctx context.Context, mctx *MiddlewareContext) {
		mctx.ResponseWriter.Header().Add("x-from-middleware", "wow")
		mctx.Next(ctx)
	})

	group := NewGroup(router, func(ctx context.Context, ba *BaseAction, next func(context.Context, *groupAction)) {
		action := &groupAction{BaseAction: ba, value: 1}
		next(ctx, action)
	})

	subgroup := NewGroup(group, func(ctx context.Context, ga *groupAction, next func(context.Context, *groupAction)) {
		ga.value += 1
		next(ctx, ga)
	})

	subgroup.Get("/hello/:name", func(ctx context.Context, c *groupAction) {
		require.Equal(t, 2, c.value)
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))
}

func TestGroup_Subrouter(t *testing.T) {
	router := New(DefaultActionFactory)

	router.Use(func(ctx context.Context, mctx *MiddlewareContext) {
		mctx.ResponseWriter.Header().Add("x-from-middleware", "wow")
		mctx.Next(ctx)
	})

	group := Subrouter(router, "/foo", func(ctx context.Context, ba *BaseAction, next func(context.Context, *groupAction)) {
		action := &groupAction{BaseAction: ba, value: 1}
		next(ctx, action)
	})

	subgroup := Subrouter(group, "/bar", func(ctx context.Context, ga *groupAction, next func(context.Context, *groupAction)) {
		ga.value += 1
		next(ctx, ga)
	})

	subgroup.Get("/hello/:name", func(ctx context.Context, c *groupAction) {
		require.Equal(t, 2, c.value)
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	subsubgroup := Subrouter(subgroup, "/baz", func(ctx context.Context, ga *groupAction, next func(context.Context, *groupAction)) {
		ga.value += 1
		next(ctx, ga)
	})

	subsubgroup.Get("/hello/:name", func(ctx context.Context, c *groupAction) {
		require.Equal(t, 3, c.value)
		c.Write([]byte(fmt.Sprintf("hello again %s", c.Params()["name"])))
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
