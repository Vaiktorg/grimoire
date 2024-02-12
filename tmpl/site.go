package tmpl

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

<<<<<<< Updated upstream:tmpl/site.go
// Site encapsulates the template engine and handles rendering of HTML templates.
// It organizes templates into partialsPath, layouts, and pages, and provides
// methods for rendering them.
//
// Example Usage:
//
//	site := NewSite("/path/to/templates", template.FuncMap{"customFunc": myFunc})
//	err := site.Render(writer, "page.layout", data)
//
// This will render the template found in "/path/to/templates/page.layout.gohtml" using
// the data provided, and write the output to the writer.
=======
//go:embed tmpl
var fs embed.FS

const tmpldir = "tmpl"

>>>>>>> Stashed changes:site/site.go
type Site struct {
	t *template.Template

	partialsPath map[string]string
	layoutPaths  map[string]string
	pagesPaths   map[string]string
}

func NewSite(funcMap template.FuncMap) *Site {
	site := &Site{
		partials: make(map[string]string),
		layouts:  make(map[string]string),
		pages:    make(map[string]string),
	}

	if err := site.parseTemplates(funcMap); err != nil {
		panic(err)
	}

	return site
}

func (s *Site) RenderPartialFunc(partialName string, data any) (template.HTML, error) {
	var buf bytes.Buffer

<<<<<<< Updated upstream:tmpl/site.go
	// Read partial template from file
	t, err := template.ParseFiles(s.partialsPath[partialName])
=======
	t, err := template.ParseFiles(s.partials[partialName])
>>>>>>> Stashed changes:site/site.go
	if err != nil {
		return "", err
	}

	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return template.HTML(buf.String()), nil
}

<<<<<<< Updated upstream:tmpl/site.go
// Render renders a template.
// w: Writer interface where output is written.
// viewName: Name of the template to render, may include layout e.g "page.layout".
// data: Data to be passed to the template.
func (s *Site) Render(w io.Writer, viewName string, data interface{}) error {
=======
func (s *Site) RenderHTML(w http.ResponseWriter, viewName string, data interface{}) {
	if err := s.render(w, viewName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Site) render(w io.Writer, viewName string, data interface{}) error {
>>>>>>> Stashed changes:site/site.go
	viewNameParts := strings.Split(viewName, ".") // 0: Page; 1: Layout

	var layoutName string

	if len(viewNameParts) == 2 {
		layoutName = viewNameParts[1]
	} else if len(viewNameParts) == 1 {
		layoutName = "index"
	}

<<<<<<< Updated upstream:tmpl/site.go
	//t := template.Must(s.t.Clone())
	//return template.Must(t.ParseFiles(append(s.partialsPath, s.layoutPaths[layoutName], s.pagesPaths[pageName])...)).ExecuteTemplate(w, layoutName, data)
	return s.t.ExecuteTemplate(w, layoutName, data)
}

// parseTemplates parses templates from rootPath and organizes them into
// layout, partial, and page categories. It also attaches any custom functions.
func (s *Site) parseTemplates(rootPath string, funcMap template.FuncMap) *Site {
	s.pagesPaths = make(map[string]string)
	s.layoutPaths = make(map[string]string)

	err := filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := strings.Split(filepath.Base(path), ".")
		if len(ext) != 3 || info.IsDir() || ext[2] != "gohtml" {
			return nil
		}
		switch ext[1] {
		case "page":
			s.pagesPaths[ext[0]] = path
		case "partial":
			s.partialsPath[ext[0]] = path
		case "layout":
			s.layoutPaths[ext[0]] = path

		}

		return nil
	})

=======
	t, err := s.t.Clone()
>>>>>>> Stashed changes:site/site.go
	if err != nil {
		return err
	}

	if s.t, err = s.t.ParseFS(fs, s.layouts[layoutName], s.pages[viewNameParts[1]]); err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	return t.ExecuteTemplate(w, layoutName, data)
}

func (s *Site) parseTemplates(funcMap template.FuncMap) error {
	s.t = template.New("").Funcs(funcMap)

	dir, err := fs.ReadDir("tmpl")
	if err != nil {
		return err
	}

	for _, file := range dir {
		path := filepath.Join(tmpldir, file.Name())

		if file.IsDir() || filepath.Ext(path) != ".gohtml" {
			continue
		}

		parts := strings.Split(strings.TrimSuffix(filepath.Base(path), ".gohtml"), ".")
		if len(parts) != 2 {
			println("invalid template path: " + path)
			continue
		}

		switch parts[1] {
		case "page":
			s.pages[parts[0]] = path
		case "partial":
			s.partials[parts[0]] = path
		case "layout":
			s.layouts[parts[0]] = path
		}
	}

	return nil
}
