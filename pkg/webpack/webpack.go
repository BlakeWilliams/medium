package webpack

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/blakewilliams/medium/pkg/router"
)

// Logger interface that is required by the middleware.
type Logger interface {
	Infof(format string, v ...any)
	Fatalf(format string, v ...any)
}

// The configuration options for the middleware
type Config struct {
	Logger      Logger
	WebpackRoot string
}

// Middleware accepts a logger and returns a middleware that can be used in
// conjunction with Router.Use.
//
// The middleware expects for webpack to be executable in the current working
// directory.
func Middleware(config Config) router.Middleware {
	logger := config.Logger
	once := sync.Once{}

	return func(c router.Action, next router.MiddlewareFunc) {
		once.Do(func() {
			go func() {
				cmd := exec.Command("npx", "webpack", "serve")
				cmd.Env = append(cmd.Env, "NODE_ENV=development")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if config.WebpackRoot != "" {
					cmd.Path = config.WebpackRoot
				}
				err := cmd.Run()

				logger.Fatalf("esbuild exited with the following error: %s", err)
			}()
		})

		if strings.HasPrefix(c.Request().URL.Path, "/assets/") {
			fileName := strings.TrimPrefix(c.Request().URL.Path, "/assets/")
			res, err := http.Get(fmt.Sprintf("http://localhost:8081/%s", fileName))

			if err != nil {
				logger.Infof("error: %s", err)
				return
			}
			defer res.Body.Close()

			if res.StatusCode == 404 {
				next(c)
			}

			if res.StatusCode < 200 || res.StatusCode > 200 {
				logger.Infof(fmt.Sprintf("Received non-2xx code when serving asset: %d", res.StatusCode))
				return
			}

			c.Response().Header().Set("Content-Type", res.Header.Get("Content-Type"))
			io.Copy(c, res.Body)
		} else {
			next(c)
		}
	}
}
