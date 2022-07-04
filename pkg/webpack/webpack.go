package webpack

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/blakewilliams/medium/pkg/router"
	"github.com/blakewilliams/medium/pkg/tell"
)

// Webpack is used to start and stop a webpack dev server instance. It can be
// used with medium/pkg/router via the Middleware method, serving assets from
// the /assets path in the target application.
type Webpack struct {
	// Path to Webpack binary
	BinPath string
	// Root dir to start webpack from
	RootDir string
	// Port for webpack to serve assets from
	Port     int
	Notifier tell.Notifier

	process *exec.Cmd
	mu      sync.Mutex
}

func New() *Webpack {
	return &Webpack{
		Notifier: tell.NullNotifier,
		Port:     9008,
	}
}

func (w *Webpack) Start(out io.Writer) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	e := w.Notifier.Start("webpack.middleware.start", tell.Payload{})
	defer e.Finish()

	if w.process != nil && !w.process.ProcessState.Exited() {
		return errors.New("process is already running")
	}

	var cmd *exec.Cmd

	if w.BinPath != "" {
		path, err := filepath.Abs(w.BinPath)
		if err != nil {
			return fmt.Errorf("could not start webpack: %w", err)
		}

		cmd = exec.Command(path, "serve", "--port", strconv.Itoa(w.Port))
	} else {
		// default to use npx
		cmd = exec.Command("npx", "webpack", "serve", "--port", strconv.Itoa(w.Port))
	}

	if w.RootDir != "" {
		path, err := filepath.Abs(w.RootDir)

		if err != nil {
			return fmt.Errorf("root dir could not be resolved: %w", err)
		}

		cmd.Dir = path
	}

	cmd.Env = append(cmd.Env, "NODE_ENV=development")
	cmd.Stdout = out
	cmd.Stderr = out
	w.process = cmd

	return cmd.Run()
}

func (w *Webpack) Stop() error {
	if w.process == nil {
		return errors.New("webpack not started")
	}

	err := w.process.Process.Signal(os.Interrupt)
	w.process.Wait()

	return err
}

// Middleware accepts a logger and returns a middleware that can be used in
// conjunction with Router.Use.
//
// The middleware expects for webpack to be executable in the current working
// directory.
//
// This is not intended for production use, just for development.
func (w *Webpack) Middleware() router.Middleware {
	return func(c router.Action, next router.MiddlewareFunc) {
		if strings.HasPrefix(c.Request().URL.Path, "/assets/") {
			// need context to coordinate timer and done statuses
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			go tryBackoff(ctx, func() {
				fileName := strings.TrimPrefix(c.Request().URL.Path, "/assets/")
				err := handleAssertRequest(c.Response(), w.Port, fileName, w.Notifier)

				if err != nil {
					var sysCallError *os.SyscallError

					switch {
					case errors.As(err, &sysCallError):
						if sysCallError.Err == syscall.ECONNREFUSED {
							// return and try again
							return
						}
					default:
					}
				}

				cancel()
			})

			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					c.ResponseWriter().WriteHeader(http.StatusInternalServerError)
					c.ResponseWriter().Write([]byte("Serving asset timed out"))
				}
			}
		} else {
			next(c)
		}
	}
}

// basic backoff functionality relying on context for cancelation
func tryBackoff(ctx context.Context, f func()) {
	attempt := 0

	for {
		if ctx.Err() != nil {
			break
		}

		if attempt > 0 {
			sleep := time.Duration(attempt * 50 * int(time.Millisecond))
			time.Sleep(sleep)
		}

		attempt++

		// if f returns true, assume success and break
		f()
	}
}

func handleAssertRequest(rw http.ResponseWriter, port int, path string, Notifier tell.Notifier) error {
	res, err := http.Get(fmt.Sprintf("http://localhost:%d/%s", port, path))
	event := Notifier.Start("webpack.serve.asset", tell.Payload{"path": path})
	defer event.Finish()

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 200 {
		event.Payload["err"] = errors.New("asset not found")
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Asset not found"))

		return nil
	}

	rw.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	io.Copy(rw, res.Body)

	return nil
}
