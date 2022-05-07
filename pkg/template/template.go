package template

import (
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"os"
	"path/filepath"
)

var TemplateUndefinedError = errors.New("template is not defined")

// Renderer is used to declare the path where templates are stored, register
// templates, and register helper functions for all templates.
type Renderer struct {
	rootPath string
	// The layout rendered by default when calling Render.
	DefaultLayout string
	HotReload     bool
	funcMap       htmlTemplate.FuncMap
	templateMap   map[string]*htmlTemplate.Template
	layoutMap     map[string]*htmlTemplate.Template
}

// Creates a new Renderer.
func New(path string) *Renderer {
	return &Renderer{
		rootPath:    path,
		funcMap:     make(htmlTemplate.FuncMap),
		templateMap: make(map[string]*htmlTemplate.Template),
		layoutMap:   make(map[string]*htmlTemplate.Template),
	}
}

// Registers a helper usable by all templates.
func (r *Renderer) Helper(name string, helper interface{}) {
	r.funcMap[name] = helper
}

// Registers a template that can be rendered. templatePath should be relative to
// the Renderer's root path.
func (r *Renderer) RegisterTemplate(templatePath string) error {
	path, err := filepath.Abs(filepath.Join(r.rootPath, templatePath))

	if err != nil {
		return fmt.Errorf("Could not register template: %e", err)
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("Could not register template: %e", err)
	}

	tmpl := htmlTemplate.Must(htmlTemplate.New(filepath.Base(path)).Funcs(r.funcMap).ParseFiles(path))
	r.templateMap[templatePath] = tmpl

	return nil
}

// Registers a layout that can be rendered. templatePath should be relative to
// the Renderer's root path.
func (r *Renderer) RegisterLayout(templatePath string) error {
	path, err := filepath.Abs(filepath.Join(r.rootPath, templatePath))

	if err != nil {
		return fmt.Errorf("Could not register template: %e", err)
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("Could not register template: %e", err)
	}

	tmpl := htmlTemplate.Must(htmlTemplate.New(filepath.Base(path)).Funcs(r.funcMap).ParseFiles(path))
	r.layoutMap[templatePath] = tmpl

	return nil
}

// Renders template with given name into the passed in io.Writer w. The data
// argument passed will be available in the template. If a DefaultLayout is
// defined on Renderer, it will also have access to the provided data.
func (r *Renderer) Render(w io.Writer, name string, data map[string]interface{}) error {
	if template, ok := r.templateMap[name]; ok {
		// TODO make hot reloading concurrency wise by extracing a template
		// object and adding a mutex
		if r.HotReload {
			err := r.RegisterTemplate(name)
			if err != nil {
				panic(err)
			}
		}

		if r.DefaultLayout != "" {
			if layout, ok := r.layoutMap[r.DefaultLayout]; ok {

				if r.HotReload {
					err := r.RegisterLayout(r.DefaultLayout)
					if err != nil {
						panic(err)
					}
				}

				buf := new(bytes.Buffer)
				if err := template.Execute(buf, data); err != nil {
					return err
				}

				data["ChildContent"] = htmlTemplate.HTML(buf.Bytes())
				if err := layout.Execute(w, data); err != nil {
					return err
				}
				delete(data, "ChildContent")
			} else {
				return fmt.Errorf("layout %s: %w", name, TemplateUndefinedError)
			}
		} else {
			if err := template.Execute(w, data); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("template %s: %w", name, TemplateUndefinedError)
	}

	return nil
}
