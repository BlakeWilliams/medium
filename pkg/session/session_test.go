package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type MyData struct {
	UserID int
	Name   string
}

func TestStoreCookie(t *testing.T) {
	verifier := NewVerifier("TheTruthIsOutThere")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := New[MyData]("session", verifier)
		session.FromRequest(r)

		session.Data.UserID = 500
		session.Data.Name = "Fox Mulder"

		err := session.Write(w)
		require.NoError(t, err)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	cookie := res.Result().Cookies()
	session := New[MyData]("session", verifier)
	err := session.FromCookie(cookie[0])

	require.NoError(t, err)
	require.Equal(t, "Fox Mulder", session.Data.Name)
	require.Equal(t, 500, session.Data.UserID)
}
func TestStoreCookie_WriteIfChanged(t *testing.T) {
	verifier := NewVerifier("TheTruthIsOutThere")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := New[MyData]("session", verifier)
		session.FromRequest(r)

		session.Data.UserID = 500
		session.Data.Name = "Fox Mulder"

		err := session.WriteIfChanged(w)
		require.NoError(t, err)
	})

	// Should set cookie since no cookies are set
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	setCookie := res.Result().Header.Get("Set-Cookie")
	require.NotEmpty(t, setCookie)

	// Second request should not have cookie set, since nothing has changed.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Cookie", setCookie)
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	setCookie = res.Result().Header.Get("Set-Cookie")
	require.Empty(t, setCookie)
}
