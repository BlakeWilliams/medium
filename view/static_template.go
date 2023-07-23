package view

import (
	"fmt"
	htmlTemplate "html/template"
	"io"
)

type staticTemplate struct {
	name         string
	renderer     *Renderer
	contents     string
	htmlTemplate *htmlTemplate.Template
}

func newStaticTemplate(renderer *Renderer, name string, contents string) (*staticTemplate, error) {
	template := &staticTemplate{name: name, renderer: renderer, contents: contents}
	err := template.compile()
	if err != nil {
		return nil, fmt.Errorf("Could not register static template: %e", err)
	}

	return template, nil
}

func (st *staticTemplate) execute(w io.Writer, data map[string]interface{}) error {
	err := st.htmlTemplate.Execute(w, data)
	if err != nil {
		return fmt.Errorf("Could not execute static template %s: %w", st.name, err)
	}

	return nil
}

func (st *staticTemplate) compile() error {
	tmpl := htmlTemplate.Must(htmlTemplate.New(st.name).Funcs(st.renderer.funcMap).Parse(st.contents))
	st.htmlTemplate = tmpl

	return nil
}
