package webpack

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/medium"
	"github.com/stretchr/testify/require"
)

func TestWebpack(t *testing.T) {
	webpack := New()
	defer func() { _ = webpack.Stop() }()
	webpack.BinPath = "./test_env/node_modules/.bin/webpack"
	webpack.RootDir = "./test_env"
	webpack.Port = 9381
	output := new(bytes.Buffer)

	err := webpack.Start(context.TODO(), output)
	require.NoError(t, err)

	r := medium.New(medium.DefaultActionCreator)
	r.Use(webpack.Middleware())

	testCases := map[string]struct {
		path              string
		expectedStatus    int
		bodyShouldContain string
	}{
		"existing file": {
			path:              "/assets/app.bundle.js",
			expectedStatus:    200,
			bodyShouldContain: "alert(",
		},
		"non-existent file": {
			path:              "/assets/missing",
			expectedStatus:    404,
			bodyShouldContain: "not found",
		},
	}
	for desc, tc := range testCases {
		t.Run(desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			res := httptest.NewRecorder()

			r.ServeHTTP(res, req)

			require.Equal(t, tc.expectedStatus, res.Result().StatusCode)

			body, err := io.ReadAll(res.Result().Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tc.bodyShouldContain)
		})
	}

	err = webpack.Stop()
	require.NoError(t, err)

	require.True(t, webpack.cmd.ProcessState.Exited())
}
