package mail

import (
	"bytes"
	"embed"
	_ "embed"
	"strconv"

	"github.com/blakewilliams/bat"
	"github.com/blakewilliams/medium"
)

//go:embed views/index.html
var indexTemplate string

//go:embed views/show.html
var showTemplate string

//go:embed views/layout.html
var layout string

//go:embed views/*
var viewFS embed.FS

func RegisterSentMailViewer[T medium.Action](medium *medium.Router[T], mailer *Mailer) {
	renderer := bat.NewEngine(bat.HTMLEscape)
	renderer.AutoRegister(viewFS, "", ".html")

	medium.Get("/_mailer", func(c T) {
		data := map[string]interface{}{
			"SentMail": mailer.SentMail,
		}

		childContent := new(bytes.Buffer)
		err := renderer.Render(childContent, "views/index.html", data)
		if err != nil {
			panic(err)
		}
		data["ChildContent"] = bat.Safe(childContent.String())

		renderer.Render(c, "views/layout.html", data)
	})

	medium.Get("/_mailer/sent/:index", func(c T) {
		strIndex := c.Params()["index"]
		index, err := strconv.Atoi(strIndex)

		if err != nil {
			panic(err)
		}

		data := map[string]interface{}{
			"Mail":  mailer.SentMail[index],
			"Index": index,
		}

		childContent := new(bytes.Buffer)
		err = renderer.Render(childContent, "views/show.html", data)
		if err != nil {
			panic(err)
		}
		data["ChildContent"] = bat.Safe(childContent.String())

		renderer.Render(c, "views/layout.html", data)
	})

	medium.Get("/_mailer/sent/:index/content/:contentIndex/body", func(c T) {
		strIndex := c.Params()["index"]
		index, err := strconv.Atoi(strIndex)
		if err != nil {
			panic(err)
		}

		strContentIndex := c.Params()["contentIndex"]
		contentIndex, err := strconv.Atoi(strContentIndex)
		if err != nil {
			panic(err)
		}

		c.Write([]byte(mailer.SentMail[index].Contents[contentIndex].Body))
	})
}
