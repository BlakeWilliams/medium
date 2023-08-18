package mail

import (
	"bytes"
	"embed"
	_ "embed"
	"strconv"

	"github.com/blakewilliams/bat"
	"github.com/blakewilliams/medium"
)

//go:embed views/*
var viewFS embed.FS

func RegisterSentMailViewer[T any](router *medium.Router[T], mailer *Mailer) {
	renderer := bat.NewEngine(bat.HTMLEscape)
	err := renderer.AutoRegister(viewFS, "", ".html")
	if err != nil {
		panic(err)
	}

	router.Get("/_mailer", func(r medium.Request[T]) {
		data := map[string]interface{}{
			"SentMail": mailer.SentMail,
		}

		childContent := new(bytes.Buffer)
		err := renderer.Render(childContent, "views/index.html", data)
		if err != nil {
			panic(err)
		}
		data["ChildContent"] = bat.Safe(childContent.String())

		_ = renderer.Render(r.Response(), "views/layout.html", data)
	})

	router.Get("/_mailer/sent/:index", func(r medium.Request[T]) {
		strIndex := r.Params()["index"]
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

		_ = renderer.Render(r.Response(), "views/layout.html", data)
	})

	router.Get("/_mailer/sent/:index/content/:contentIndex/body", func(r medium.Request[T]) {
		strIndex := r.Params()["index"]
		index, err := strconv.Atoi(strIndex)
		if err != nil {
			panic(err)
		}

		strContentIndex := r.Params()["contentIndex"]
		contentIndex, err := strconv.Atoi(strContentIndex)
		if err != nil {
			panic(err)
		}

		r.Response().Write([]byte(mailer.SentMail[index].Contents[contentIndex].Body))
	})
}
