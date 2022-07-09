package mail

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/blakewilliams/medium/pkg/router"
	"github.com/blakewilliams/medium/pkg/view"
	"github.com/stretchr/testify/require"
)

func TestSentViewer_Index(t *testing.T) {
	r := router.New(router.DefaultActionFactory)
	renderer := view.New(os.DirFS("/"))
	renderer.RegisterStaticTemplate("index", "welcome!")

	mailer := New(&FakeDeliverer{}, renderer)
	mailer.DevMode = true
	mailer.From = "noreply@bar.net"

	RegisterSentMailViewer(r, mailer)

	req := httptest.NewRequest("GET", "/_mailer", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Equal(t, 200, res.Code)
	require.Contains(t, res.Body.String(), "No mail has been sent")

	err := mailer.Send("index", "foo@bar.net", "Welcome!", map[string]any{})
	require.NoError(t, err)

	req = httptest.NewRequest("GET", "/_mailer", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Contains(t, res.Body.String(), `<a href="/_mailer/sent/0">`)
	require.Contains(t, res.Body.String(), "foo@bar.net")
	require.Contains(t, res.Body.String(), "Welcome!")
	require.Contains(t, res.Body.String(), "noreply@bar.net")

	r.ServeHTTP(res, req)
}

func TestSentViewer_Show(t *testing.T) {
	r := router.New(router.DefaultActionFactory)
	renderer := view.New(os.DirFS("/"))
	renderer.RegisterStaticTemplate("index", "welcome!")

	mailer := New(&FakeDeliverer{}, renderer)
	mailer.DevMode = true
	mailer.From = "noreply@bar.net"

	RegisterSentMailViewer(r, mailer)

	mailer.Send("index", "foo@bar.net", "Welcome!", map[string]any{})

	req := httptest.NewRequest("GET", "/_mailer/sent/0", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Contains(t, res.Body.String(), `<iframe src="/_mailer/sent/0/body">`)
	require.Contains(t, res.Body.String(), "foo@bar.net")
	require.Contains(t, res.Body.String(), "Welcome!")
	require.Contains(t, res.Body.String(), "noreply@bar.net")

	r.ServeHTTP(res, req)
}

func TestSentViewer_Body(t *testing.T) {
	r := router.New(router.DefaultActionFactory)
	renderer := view.New(os.DirFS("/"))
	renderer.RegisterStaticTemplate("index", "welcome!")

	mailer := New(&FakeDeliverer{}, renderer)
	mailer.DevMode = true
	mailer.From = "noreply@bar.net"

	RegisterSentMailViewer(r, mailer)

	mailer.Send("index", "foo@bar.net", "Welcome!", map[string]any{})

	req := httptest.NewRequest("GET", "/_mailer/sent/0/body", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Equal(t, res.Body.String(), "welcome!")

	r.ServeHTTP(res, req)
}
