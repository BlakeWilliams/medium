package apptest

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponseBody_MultiReadable(t *testing.T) {
	rec := &httptest.ResponseRecorder{
		Body: bytes.NewBuffer([]byte("omg wow")),
	}

	res := Response{RawResponse: rec}

	require.Equal(t, "omg wow", res.Body())
	require.Equal(t, "omg wow", res.Body())
}
