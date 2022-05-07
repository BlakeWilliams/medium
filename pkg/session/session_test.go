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

func TestSessionCookie(t *testing.T) {
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
