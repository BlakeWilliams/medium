package mail

import (
	"context"
	"embed"
	_ "embed"
	"strconv"

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/pkg/view"
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
	renderer := view.New(viewFS)
	renderer.RegisterTemplate("views/index.html")
	renderer.RegisterTemplate("views/show.html")

	renderer.RegisterLayout("views/layout.html")
	renderer.DefaultLayout = "views/layout.html"

	medium.Get("/_mailer", func(ctx context.Context, c T) {
		renderer.Render(c, "views/index.html", map[string]interface{}{
			"SentMail": mailer.SentMail,
		})
	})

	medium.Get("/_mailer/sent/:index", func(ctx context.Context, c T) {
		strIndex := c.Params()["index"]
		index, err := strconv.Atoi(strIndex)

		if err != nil {
			panic(err)
		}

		err = renderer.Render(c, "views/show.html", map[string]interface{}{
			"Mail":  mailer.SentMail[index],
			"Index": index,
		})

		if err != nil {
			panic(err)
		}
	})

	medium.Get("/_mailer/sent/:index/body", func(ctx context.Context, c T) {
		strIndex := c.Params()["index"]
		index, err := strconv.Atoi(strIndex)

		if err != nil {
			panic(err)
		}

		c.Write([]byte(mailer.SentMail[index].Body))
	})
}
