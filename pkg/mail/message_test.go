package mail

import (
	"io"
	"mime/multipart"
	"regexp"
	"strings"
	"testing"

	"github.com/blakewilliams/bat"
	"github.com/stretchr/testify/require"
)

var boundaryContentRegex = regexp.MustCompile(`multipart/mixed; boundary=(.+)`)

func TestMessage_Multipart(t *testing.T) {
	renderer := bat.NewEngine(bat.HTMLEscape)
	err := renderer.AutoRegister(fixtureViewFS, "fixtures", ".html")
	require.NoError(t, err)
	err = renderer.AutoRegister(fixtureViewFS, "fixtures", ".txt")
	require.NoError(t, err)

	fakeDeliverer := &FakeDeliverer{}
	mailer := New(fakeDeliverer, renderer)
	mailer.DevMode = true

	msg := mailer.NewMessage("Hello!", "fox@mulder.net")
	err = msg.Multipart([]string{"welcome.html", "welcome.txt"}, map[string]any{"Name": "Fox Mulder"})
	require.NoError(t, err)

	require.Regexp(t, boundaryContentRegex, msg.ContentType)
	boundary := boundaryContentRegex.FindStringSubmatch(msg.ContentType)[1]

	r := multipart.NewReader(strings.NewReader(msg.Body), boundary)
	html, err := r.NextPart()
	require.NoError(t, err)

	require.Equal(t, "text/html; charset=utf-8", html.Header.Get("Content-Type"))
	htmlBody, _ := io.ReadAll(html)
	require.Equal(t, "<h1>Welcome, Fox Mulder!</h1>\n", string(htmlBody))

	text, err := r.NextPart()
	require.NoError(t, err)

	require.Equal(t, "text/plain", text.Header.Get("Content-Type"))
	textBody, _ := io.ReadAll(text)
	require.Equal(t, "Welcome, Fox Mulder!\n", string(textBody))

	_, err = r.NextPart()
	require.ErrorIs(t, err, io.EOF)
}

func TestMessage_Template_PlainText(t *testing.T) {
	renderer := bat.NewEngine(bat.HTMLEscape)
	err := renderer.AutoRegister(fixtureViewFS, "fixtures", ".html")
	require.NoError(t, err)
	err = renderer.AutoRegister(fixtureViewFS, "fixtures", ".txt")
	require.NoError(t, err)

	fakeDeliverer := &FakeDeliverer{}
	mailer := New(fakeDeliverer, renderer)
	mailer.DevMode = true

	msg := mailer.NewMessage("Hello!", "fox@mulder.net")
	err = msg.Template("welcome.txt", map[string]any{"Name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "text/plain", msg.ContentType)
	require.Equal(t, "Welcome, Fox Mulder!\n", msg.Body)
}
