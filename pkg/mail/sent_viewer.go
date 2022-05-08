package mail

import (
	_ "embed"
	"strconv"

	"github.com/blakewilliams/medium/pkg/router"
	"github.com/blakewilliams/medium/pkg/template"
)

//go:embed views/index.html
var indexTemplate string

//go:embed views/show.html
var showTemplate string

//go:embed views/layout.html
var layout string

func RegisterSentMailViewer[T router.Action](router *router.Router[T], mailer *Mailer) {
	renderer := template.New("")
	renderer.RegisterStaticTemplate("index", indexTemplate)
	renderer.RegisterStaticTemplate("show", showTemplate)

	renderer.RegisterStaticLayout("layout", layout)
	renderer.DefaultLayout = "layout"

	router.Get("/_mailer", func(c T) {
		renderer.Render(c, "index", map[string]interface{}{
			"SentMail": mailer.SentMail,
		})
	})

	router.Get("/_mailer/sent/:index", func(c T) {
		strIndex := c.Params()["index"]
		index, err := strconv.Atoi(strIndex)

		if err != nil {
			panic(err)
		}

		err = renderer.Render(c, "show", map[string]interface{}{
			"Mail":  mailer.SentMail[index],
			"Index": index,
		})

		if err != nil {
			panic(err)
		}
	})

	router.Get("/_mailer/sent/:index/body", func(c T) {
		strIndex := c.Params()["index"]
		index, err := strconv.Atoi(strIndex)

		if err != nil {
			panic(err)
		}

		c.Write([]byte(mailer.SentMail[index].Body))
	})
}
