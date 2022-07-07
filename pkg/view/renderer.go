package view

import (
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"io/fs"
	"strings"
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
	FS            fs.FS
}

type renderable interface {
	execute(w io.Writer, data map[string]interface{}) error
	compile() error
}

// Creates a new Renderer. A fs.FS is accepted which allows usage of go:embed
// to embed all templates into the binary as a virtual file system.
//
// If not using embed, os.DirFS(path) can be passed to use the real filesystem
// with path acting as the root directory.
func New(rootFS fs.FS) *Renderer {
	return &Renderer{
		funcMap:     make(htmlTemplate.FuncMap),
		templateMap: make(map[string]renderable),
		layoutMap:   make(map[string]renderable),
		FS:          rootFS,
	}
}

// Registers a helper usable by all templates.
func (r *Renderer) Helper(name string, helper interface{}) {
	r.funcMap[name] = helper
}

// Registers all templates in the given directory. The layouts subdirectory, if
// present, register all files within it as layouts.
func (r *Renderer) AutoRegister() error {
	err := fs.WalkDir(r.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(path, "layouts") {
			r.RegisterLayout(path)
		} else {
			r.RegisterTemplate(path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("Could not register templates in %s: %w", r.rootPath, err)
	}

	return nil
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
