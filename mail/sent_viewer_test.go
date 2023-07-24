package mail

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/blakewilliams/bat"
	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/view"
	"github.com/stretchr/testify/require"
)

func TestSentViewer_Index(t *testing.T) {
	r := medium.New(medium.DefaultActionCreator)
	mailer := New(&FakeDeliverer{}, sentRenderer(t))
	mailer.DevMode = true
	mailer.From = "noreply@bar.net"

	RegisterSentMailViewer(r, mailer)

	req := httptest.NewRequest("GET", "/_mailer", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Equal(t, 200, res.Code)
	require.Contains(t, res.Body.String(), "No mail has been sent")

	msg := mailer.NewMessage("Welcome!", "foo@bar.net")
	err := msg.Template("index.html", nil)
	require.NoError(t, err)

	err = mailer.Send(context.Background(), msg)
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
	r := medium.New(medium.DefaultActionCreator)
	renderer := view.New(os.DirFS("/"))
	err := renderer.RegisterStaticTemplate("index", "welcome!")
	require.NoError(t, err)

	mailer := New(&FakeDeliverer{}, sentRenderer(t))
	mailer.DevMode = true
	mailer.From = "noreply@bar.net"

	RegisterSentMailViewer(r, mailer)

	msg := mailer.NewMessage("Welcome!", "foo@bar.net")
	err = msg.Template("index.html", nil)
	require.NoError(t, err)

	err = mailer.Send(context.Background(), msg)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/_mailer/sent/0", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Contains(t, res.Body.String(), `<iframe src="/_mailer/sent/0/content/0/body">`)
	require.Contains(t, res.Body.String(), "foo@bar.net")
	require.Contains(t, res.Body.String(), "Welcome!")
	require.Contains(t, res.Body.String(), "noreply@bar.net")

	r.ServeHTTP(res, req)
}

func TestSentViewer_Body(t *testing.T) {
	r := medium.New(medium.DefaultActionCreator)
	mailer := New(&FakeDeliverer{}, sentRenderer(t))
	mailer.DevMode = true
	mailer.From = "noreply@bar.net"

	RegisterSentMailViewer(r, mailer)

	msg := mailer.NewMessage("Welcome!", "foo@bar.net")
	err := msg.Template("index.html", nil)
	require.NoError(t, err)

	err = mailer.Send(context.Background(), msg)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/_mailer/sent/0/content/0/body", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	require.Equal(t, res.Body.String(), "welcome!")

	r.ServeHTTP(res, req)
}

func sentRenderer(tb testing.TB) *bat.Engine {
	engine := bat.NewEngine(bat.HTMLEscape)
	err := engine.Register("index.html", "welcome!")
	require.NoError(tb, err)

	return engine
}
