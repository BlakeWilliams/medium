package slashnormalizer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMiddleware_DoubleSlash(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "//hello", nil)
	w := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	Middleware(w, r, next)
	require.False(t, called)
	require.Equal(t, http.StatusMovedPermanently, w.Code)
	require.Equal(t, "/hello", w.Header().Get("Location"))
}

func TestMiddleware_Valid(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	Middleware(w, r, next)
	require.True(t, called)
	require.Equal(t, http.StatusOK, w.Code)
}
