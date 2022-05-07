package template

import (
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"io"
)

var TemplateUndefinedError = errors.New("template is not defined")

// Renderer is used to declare the path where templates are stored, register
// templates, and register helper functions for all templates.
type Renderer struct {
	// The layout rendered by default when calling Render.
	DefaultLayout string
	HotReload     bool
	rootPath      string
	funcMap       htmlTemplate.FuncMap
	templateMap   map[string]renderable
	layoutMap     map[string]renderable
}

type renderable interface {
	execute(w io.Writer, data map[string]interface{}) error
	compile() error
}

// Creates a new Renderer.
func New(path string) *Renderer {
	return &Renderer{
		rootPath:    path,
		funcMap:     make(htmlTemplate.FuncMap),
		templateMap: make(map[string]renderable),
		layoutMap:   make(map[string]renderable),
	}
}

// Registers a helper usable by all templates.
func (r *Renderer) Helper(name string, helper interface{}) {
	r.funcMap[name] = helper
}

// Registers a template that can be rendered. templatePath should be relative to
// the Renderer's root path.
func (r *Renderer) RegisterTemplate(templatePath string) error {
	tmpl, err := newfsTemplate(r, templatePath)
	if err != nil {
		return err
	}

	r.templateMap[templatePath] = tmpl

	return nil
}

// Registers a template using the passed in contents as the template source.
func (r *Renderer) RegisterStaticTemplate(name string, contents string) error {
	tmpl, err := newStaticTemplate(r, name, contents)

	if err != nil {
		return err
	}

	r.templateMap[name] = tmpl

	return nil
}

// Registers a layout using the passed in contents as the template source.
func (r *Renderer) RegisterStaticLayout(name string, contents string) error {
	tmpl, err := newStaticTemplate(r, name, contents)

	if err != nil {
		return err
	}

	r.layoutMap[name] = tmpl

	return nil
}

// Registers a layout that can be rendered. templatePath should be relative to
// the Renderer's root path.
func (r *Renderer) RegisterLayout(templatePath string) error {
	tmpl, err := newfsTemplate(r, templatePath)
	if err != nil {
		return err
	}

	r.layoutMap[templatePath] = tmpl

	return nil
}

// Renders template with given name into the passed in io.Writer w. The data
// argument passed will be available in the template. If a DefaultLayout is
// defined on Renderer, it will also have access to the provided data.
func (r *Renderer) Render(w io.Writer, name string, data map[string]interface{}) error {
	// This isn't as efficient as it could be, but it makes the code easier to
	// read for now.
	content, err := r.RenderString(name, data)
	if err != nil {
		return err
	}

	if r.DefaultLayout != "" {
		layout, ok := r.layoutMap[r.DefaultLayout]
		if !ok {
			return fmt.Errorf("layout %s: %w", r.DefaultLayout, TemplateUndefinedError)
		}

		data["ChildContent"] = content
		if err := layout.execute(w, data); err != nil {
			return err
		}
		delete(data, "ChildContent")
	} else {
		_, err := w.Write([]byte(content))
		if err != nil {
			return fmt.Errorf("Could not write template to writer: %w", err)
		}
	}

	return nil
}

func (r *Renderer) RenderString(name string, data map[string]interface{}) (htmlTemplate.HTML, error) {
	template, ok := r.templateMap[name]

	if !ok {
		return "", fmt.Errorf("template %s: %w", name, TemplateUndefinedError)
	}

	buf := new(bytes.Buffer)
	if err := template.execute(buf, data); err != nil {
		return "", err
	}

	return htmlTemplate.HTML(buf.String()), nil
}
