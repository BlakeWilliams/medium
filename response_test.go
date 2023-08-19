package medium

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedirect(t *testing.T) {
	res := Redirect("https://google.com")
	require.Equal(t, 302, res.Status())
	require.Equal(t, "https://google.com", res.Header().Get("Location"))
}

func TestResponseBuilder(t *testing.T) {
	res := ResponseBuilder{}
	res.WriteStatus(200)
	res.Header().Add("x-test", "test")
	res.WriteString("hello")

	require.Equal(t, 200, res.Status())
	require.Equal(t, "test", res.Header().Get("x-test"))

	body, _ := io.ReadAll(res.Body())
	require.Equal(t, "hello", string(body))
}

func TestResponseBuilder_MultiWrite(t *testing.T) {
	res := NewResponse()

	res.Write([]byte("hello"))
	res.WriteString(" world")

	body, _ := io.ReadAll(res.Body())
	require.Equal(t, "hello world", string(body))
}
