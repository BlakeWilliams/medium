package mail

import (
	"embed"
	"testing"

	"github.com/blakewilliams/medium/pkg/view"
	"github.com/stretchr/testify/require"
)

type FakeDeliverer struct {
	deliveries int
}

func (f *FakeDeliverer) SendMail(addr string, addrs []string, subject string, msg []byte) error {
	f.deliveries++
	return nil
}

//go:embed fixtures/*
var fixtureViewFS embed.FS

func TestMail_SentMail(t *testing.T) {
	renderer := view.New(fixtureViewFS)
	err := renderer.RegisterTemplate("fixtures/welcome.html")
	require.NoError(t, err)

	fakeDeliverer := &FakeDeliverer{}
	mailer := New(fakeDeliverer, renderer)
	mailer.DevMode = true
	mailer.From = "noreply@bar.net"

	require.NoError(t, err)

	err = mailer.Send("fixtures/welcome.html", "foo@bar.net", "Hello!", map[string]interface{}{"Name": "Walter Skinner"})
	require.NoError(t, err)

	require.Equal(t, 1, len(mailer.SentMail))

	mail := mailer.SentMail[0]
	require.Equal(t, "<h1>Welcome, Walter Skinner!</h1>\n", mail.Body)
	require.Equal(t, "foo@bar.net", mail.To[0])
	require.Equal(t, "Hello!", mail.Subject)
	require.Equal(t, "noreply@bar.net", mail.From)
	require.NotNil(t, mail.SentAt)

	require.Equal(t, 0, fakeDeliverer.deliveries)
}

func TestMail_SentMail_DevModeFalse(t *testing.T) {
	renderer := view.New(fixtureViewFS)
	err := renderer.RegisterTemplate("fixtures/welcome.html")
	require.NoError(t, err)

	fakeDeliverer := &FakeDeliverer{}
	mailer := New(fakeDeliverer, renderer)
	mailer.DevMode = false
	mailer.From = "noreply@bar.net"
	require.NoError(t, err)

	err = mailer.Send("fixtures/welcome.html", "foo@bar.net", "Hello!", map[string]interface{}{"Name": "Walter Skinner"})
	require.NoError(t, err)

	require.Equal(t, 0, len(mailer.SentMail))
	require.Equal(t, 1, fakeDeliverer.deliveries)
}
