package template

import (
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
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
	templateMap   map[string]*registeredTemplate
	layoutMap     map[string]*registeredTemplate
}

type registeredTemplate struct {
	mu           sync.Mutex
	path         string
	renderer     *Renderer
	htmlTemplate *htmlTemplate.Template
}

func newRegisteredTemplate(renderer *Renderer, path string) (*registeredTemplate, error) {
	rt := &registeredTemplate{path: path, renderer: renderer}
	err := rt.compile()

	return rt, err
}

func (rt *registeredTemplate) execute(w io.Writer, data map[string]interface{}) error {
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

func (rt *registeredTemplate) compile() error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	path, err := filepath.Abs(filepath.Join(rt.renderer.rootPath, rt.path))

	if err != nil {
		return fmt.Errorf("Could not register template: %e", err)
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("Could not register template: %e", err)
	}

	tmpl := htmlTemplate.Must(htmlTemplate.New(filepath.Base(path)).Funcs(rt.renderer.funcMap).ParseFiles(path))
	rt.htmlTemplate = tmpl

	return nil
}

// Creates a new Renderer.
func New(path string) *Renderer {
	return &Renderer{
		rootPath:    path,
		funcMap:     make(htmlTemplate.FuncMap),
		templateMap: make(map[string]*registeredTemplate),
		layoutMap:   make(map[string]*registeredTemplate),
	}
}

// Registers a helper usable by all templates.
func (r *Renderer) Helper(name string, helper interface{}) {
	r.funcMap[name] = helper
}

// Registers a template that can be rendered. templatePath should be relative to
// the Renderer's root path.
func (r *Renderer) RegisterTemplate(templatePath string) error {
	tmpl, err := newRegisteredTemplate(r, templatePath)
	if err != nil {
		return err
	}

	r.templateMap[templatePath] = tmpl

	return nil
}

// Registers a layout that can be rendered. templatePath should be relative to
// the Renderer's root path.
func (r *Renderer) RegisterLayout(templatePath string) error {
	tmpl, err := newRegisteredTemplate(r, templatePath)
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
