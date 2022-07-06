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

	w.out.Write(text)
	w.out.Write([]byte("\n"))
}
