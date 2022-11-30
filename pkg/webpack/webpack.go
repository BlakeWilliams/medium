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

	"github.com/blakewilliams/medium"
	"github.com/blakewilliams/medium/pkg/mlog"
)

// Webpack is used to start and stop a webpack dev server instance. It can be
// used with medium/pkg/medium via the Middleware method, serving assets from
// the /assets path in the target application.
type Webpack struct {
	// Path to Webpack binary
	BinPath string
	// Root dir to start webpack from
	RootDir string
	// Port for webpack to serve assets from
	Port int

	cmd *exec.Cmd
	mu  sync.Mutex
}

func New() *Webpack {
	return &Webpack{
		Port: 9008,
	}
}

func (w *Webpack) Start(ctx context.Context, out io.Writer) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cmd != nil && !w.cmd.ProcessState.Exited() {
		return errors.New("process is already running")
	}

	var cmd *exec.Cmd

	if w.BinPath != "" {
		path, err := filepath.Abs(w.BinPath)
		if err != nil {
			return fmt.Errorf("could not start webpack: %w", err)
		}

		cmd = exec.CommandContext(ctx, path, "serve", "--port", strconv.Itoa(w.Port))
	} else {
		// default to use npx
		cmd = exec.CommandContext(ctx, "npx", "webpack", "serve", "--port", strconv.Itoa(w.Port))
	}

	if w.RootDir != "" {
		path, err := filepath.Abs(w.RootDir)

		if err != nil {
			return fmt.Errorf("root dir could not be resolved: %w", err)
		}

		cmd.Dir = path
	}
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "NODE_ENV=development")

	cmd.Stdout = out
	cmd.Stderr = out
	w.cmd = cmd

	mlog.Debug(ctx, "Starting webpack server", mlog.Fields{"port": w.Port})

	return cmd.Start()
}

func (w *Webpack) Wait() error {
	if w.cmd == nil {
		return errors.New("webpack not running")
	}

	return w.cmd.Wait()
}

func (w *Webpack) Stop() error {
	if w.cmd == nil || w.cmd.Process == nil {
		return errors.New("webpack not started")
	}

	err := w.cmd.Process.Signal(os.Interrupt)
	w.cmd.Wait()

	return err
}

// Middleware accepts a logger and returns a middleware that can be used in
// conjunction with medium.Use.
//
// The middleware expects for webpack to be executable in the current working
// directory.
//
// This is not intended for production use, just for development.
func (w *Webpack) Middleware() medium.MiddlewareFunc {
	return func(ctx context.Context, r *http.Request, rw http.ResponseWriter, next medium.NextMiddleware) {
		start := time.Now()

		if strings.HasPrefix(r.URL.Path, "/assets/") {
			if w.cmd == nil || w.cmd.Process == nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte("Webpack not running"))
				return
			}

			// need context to coordinate timer and done statuses
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			go tryBackoff(ctx, func() {
				fileName := strings.TrimPrefix(r.URL.Path, "/assets/")
				err := handleAssertRequest(ctx, rw, w.Port, fileName)

				if err != nil {
					var sysCallError *os.SyscallError

					switch {
					case errors.As(err, &sysCallError):
						if sysCallError.Err == syscall.ECONNREFUSED {
							// return and try again
							return
						}
					default:
						mlog.Error(ctx, "webpack could not serve asset", mlog.Fields{"error": err, "path": r.URL.Path})
					}
				}

				cancel()
			})

			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					mlog.Error(
						ctx,
						"Webpack asset request failed",
						mlog.Fields{"path": r.URL.Path, "error": ctx.Err()},
					)
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write([]byte("Serving asset timed out"))
					return
				}

				mlog.Debug(ctx, "webpack asset served", mlog.Fields{"path": r.URL.Path, "duration": time.Since(start).String()})
			}
		} else {
			next(ctx, r, rw)
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

func handleAssertRequest(ctx context.Context, rw http.ResponseWriter, port int, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost:%d/%s", port, path), nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 200 {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("Asset not found"))

		return nil
	}

	rw.Header().Set("Content-Type", res.Header.Get("Content-Type"))
	io.Copy(rw, res.Body)

	return nil
}
