package rescue

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/medium"
	"github.com/stretchr/testify/require"
)

func TestRescue(t *testing.T) {
	r := medium.New(medium.DefaultActionFactory)

	handler := func(ctx context.Context, r *http.Request, rw http.ResponseWriter, err error) {
		rw.Write([]byte("oh no!"))
	}

	r.Use(Middleware(handler))
	r.Get("/", func(ctx context.Context, ac *medium.BaseAction) {
		panic("oh no!")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	require.Equal(t, "oh no!", res.Body.String())
}
