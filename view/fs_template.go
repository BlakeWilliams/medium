package view

import (
	"fmt"
	htmlTemplate "html/template"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
)

type fsTemplate struct {
	mu           sync.Mutex
	path         string
	renderer     *Renderer
	htmlTemplate *htmlTemplate.Template
}

func newfsTemplate(renderer *Renderer, path string) (*fsTemplate, error) {
	rt := &fsTemplate{path: path, renderer: renderer}
	err := rt.compile()

	return rt, err
}

func (rt *fsTemplate) execute(w io.Writer, data map[string]interface{}) error {
	if rt.renderer.HotReload {
		err := rt.compile()

		if err != nil {
			return fmt.Errorf("Could not hot compile template %s: %w", rt.path, err)
		}
	}

	err := rt.htmlTemplate.Execute(w, data)

	if err != nil {
		return fmt.Errorf("Could not execute template %s: %w", rt.path, err)
	}

	return nil
}

func (rt *fsTemplate) compile() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if _, err := fs.Stat(rt.renderer.FS, rt.path); err != nil {
		return fmt.Errorf("Could not register template: %e", err)
	}

	tmpl := htmlTemplate.Must(htmlTemplate.New(filepath.Base(rt.path)).Funcs(rt.renderer.funcMap).ParseFS(rt.renderer.FS, rt.path))
	rt.htmlTemplate = tmpl

	return nil
}
