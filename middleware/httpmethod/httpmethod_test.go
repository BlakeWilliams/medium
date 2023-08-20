package httpmethod

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/medium"
	"github.com/stretchr/testify/require"
)

func TestRewrite(t *testing.T) {
	r := medium.New(medium.WithNoData)

	r.Use(RewriteMiddleware)
	r.Delete("/", func(ctx context.Context, r *medium.Request[medium.NoData]) medium.Response { return medium.OK() })

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Result().StatusCode)
}

func Test_RewritePost(t *testing.T) {
	r := medium.New(medium.WithNoData)

	r.Use(RewriteMiddleware)
	r.Delete("/", func(ctx context.Context, ac *medium.Request[medium.NoData]) medium.Response { return medium.OK() })

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	require.Equal(t, http.StatusNotFound, res.Result().StatusCode)
}
