package mlog

import (
	"io"
	"sync"
)

type writer struct {
	out io.Writer
	mu  sync.Mutex
}

func (w *writer) Print(text []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, _ = w.out.Write(text)
	_, _ = w.out.Write([]byte("\n"))
}
