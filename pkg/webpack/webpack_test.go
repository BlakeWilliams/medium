package webpack

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/medium/pkg/router"
	"github.com/stretchr/testify/require"
)

func TestWebpack(t *testing.T) {
	webpack := New()
	defer webpack.Stop()
	webpack.RootDir = "./test_env"
	webpack.Port = 9381
	var output []byte

	done := make(chan struct{})
	go func(done chan struct{}) {
		defer close(done)
		buf := bytes.NewBuffer(output)
		err := webpack.Start(buf)
		require.NoError(t, err, buf.String())
	}(done)

	r := router.New(router.DefaultActionFactory)
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

	err := webpack.Stop()
	require.NoError(t, err)

	<-done
	require.True(t, webpack.process.ProcessState.Exited())
}
