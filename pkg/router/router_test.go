package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {
	router := New(DefaultControllerFactory)

	router.Get("/hello/:name", func(c BaseController) {
		c.Write([]byte(fmt.Sprintf("hello %s", c.params["name"])))
	})

	req := httptest.NewRequest(http.MethodGet, "/hello/Fox%20Mulder", nil)
	rw := httptest.NewRecorder()

	router.Run(rw, req)

	require.Equal(t, "hello Fox Mulder", rw.Body.String())
}
