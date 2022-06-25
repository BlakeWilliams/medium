package rescue

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/medium/pkg/router"
	"github.com/stretchr/testify/require"
)

func TestRescue(t *testing.T) {
	r := router.New(router.DefaultActionFactory)

	handler := func(ac router.Action, err error) {
		ac.Write([]byte("oh no!"))
	}

	r.Use(Middleware(handler))
	r.Get("/", func(ac *router.BaseAction) {
		panic("oh no!")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	r.Run(res, req)

	require.Equal(t, "oh no!", res.Body.String())
}
