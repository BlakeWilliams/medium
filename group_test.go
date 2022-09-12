package medium

import (
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

	router.Use(func(a Action, next MiddlewareFunc) { next(a) })

	group := NewGroup(router, func(ba *BaseAction, next func(*groupAction)) {
		action := &groupAction{BaseAction: ba}
		next(action)
	})

	for name, tc := range testCases {
		path := reflect.ValueOf("/")
		handler := reflect.ValueOf(func(c *groupAction) {
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

	router.Use(func(a Action, next MiddlewareFunc) {
		a.Response().Header().Add("x-from-middleware", "wow")
		next(a)
	})

	group := NewGroup(router, func(ba *BaseAction, next func(*groupAction)) {
		action := &groupAction{BaseAction: ba}
		next(action)
	})
	group.Get("/hello/:name", func(c *groupAction) {
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

	router.Use(func(a Action, next MiddlewareFunc) {
		a.Response().Header().Add("x-from-middleware", "wow")
		next(a)
	})

	group := NewGroup(router, func(ba *BaseAction, next func(*groupAction)) {
		action := &groupAction{BaseAction: ba, value: 1}
		next(action)
	})

	subgroup := NewGroup(group, func(ga *groupAction, next func(*groupAction)) {
		ga.value += 1
		next(ga)
	})

	subgroup.Get("/hello/:name", func(c *groupAction) {
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

	router.Use(func(a Action, next MiddlewareFunc) {
		a.Response().Header().Add("x-from-middleware", "wow")
		next(a)
	})

	group := Subrouter(router, "/foo", func(ba *BaseAction, next func(*groupAction)) {
		action := &groupAction{BaseAction: ba, value: 1}
		next(action)
	})

	subgroup := Subrouter(group, "/bar", func(ga *groupAction, next func(*groupAction)) {
		ga.value += 1
		next(ga)
	})

	subgroup.Get("/hello/:name", func(c *groupAction) {
		require.Equal(t, 2, c.value)
		c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
	})

	subsubgroup := Subrouter(subgroup, "/baz", func(ga *groupAction, next func(*groupAction)) {
		ga.value += 1
		next(ga)
	})

	subsubgroup.Get("/hello/:name", func(c *groupAction) {
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
