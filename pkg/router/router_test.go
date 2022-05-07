package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	router := New(NewContext)

	router.Use(func(c *BaseContext, next HandlerFunc[*BaseContext]) {
		c.Response().Header().Add("x-from-middleware", "wow")
		next(c)
	})

	router.Controller("/", func(c *Controller[*BaseContext]) {
		c.Get("/", func(c *BaseContext) {
			c.Write([]byte("home page!"))
		})
	})

	router.Controller("/hello", func(c *Controller[*BaseContext]) {
		c.Before(func(c *BaseContext, next HandlerFunc[*BaseContext]) {
			c.Response().Header().Add("Content-Type", "text/plain")
			next(c)
		})

		c.Get("/:name", func(c *BaseContext) {
			c.Write([]byte(fmt.Sprintf("hello %s", c.Params()["name"])))
		})
	})

	router.Controller("/goodbye", func(c *Controller[*BaseContext]) {
		c.Before(func(c *BaseContext, next HandlerFunc[*BaseContext]) {
			c.Response().Header().Add("Content-Type", "text/plain")
			next(c)
		})

		c.Before(func(c *BaseContext, next HandlerFunc[*BaseContext]) {
			c.Response().Header().Add("x-custom", "thetruthisouthere")
			next(c)
		})

		c.Get("/:name", func(c *BaseContext) {
			c.Write([]byte(fmt.Sprintf("goodbye %s", c.Params()["name"])))
		})

		c.Before(func(c *BaseContext, next HandlerFunc[*BaseContext]) {
			c.Response().Header().Add("x-missing", "conspiracy")
			next(c)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	router.Run(rw, req)

	require.Equal(t, "home page!", rw.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw = httptest.NewRecorder()

	router.Run(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
	require.Equal(t, "text/plain", rw.Header().Get("content-type"))
	require.Equal(t, "wow", rw.Header().Get("x-from-middleware"))

	req = httptest.NewRequest(http.MethodGet, "/goodbye/Dana%20Scully", nil)
	rw = httptest.NewRecorder()

	router.Run(rw, req)

	require.Equal(t, "goodbye Dana Scully", rw.Body.String())
	require.Equal(t, "text/plain", rw.Header().Get("content-type"))
	require.Equal(t, "thetruthisouthere", rw.Header().Get("x-custom"))
	// Before action does not get called for the GET /:name action due to it
	// being called after the c.Get call
	require.Equal(t, "", rw.Header().Get("x-missing"))
}
