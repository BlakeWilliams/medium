package template

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicTemplate(t *testing.T) {
	renderer := New("./fixtures")
	err := renderer.RegisterTemplate("hello.tmpl.html")
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = renderer.Render(buf, "hello.tmpl.html", map[string]interface{}{"Name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello Fox Mulder</h1>\n", buf.String())
}

func TestBasicTemplateWithHelper(t *testing.T) {
	renderer := New("./fixtures")
	renderer.Helper("upcase", func(word string) string {
		return strings.ToUpper(word)
	})
	err := renderer.RegisterTemplate("funcs.tmpl.html")
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = renderer.Render(buf, "funcs.tmpl.html", map[string]interface{}{"Name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "<h1>Hello FOX MULDER</h1>\n", buf.String())
}

func TestTemplateWithLayout(t *testing.T) {
	renderer := New("./fixtures")
	renderer.Helper("upcase", func(word string) string {
		return strings.ToUpper(word)
	})
	err := renderer.RegisterLayout("layout.tmpl.html")
	require.NoError(t, err)

	renderer.DefaultLayout = "layout.tmpl.html"

	err = renderer.RegisterTemplate("funcs.tmpl.html")
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = renderer.Render(buf, "funcs.tmpl.html", map[string]interface{}{"Name": "Fox Mulder"})
	require.NoError(t, err)

	require.Equal(t, "<html>\n<body>\n  <h1>Hello FOX MULDER</h1>\n\n</body>\n</html>\n", buf.String())
}
