package site

import (
	"bytes"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

/*
	 Site encapsulates the template engine and handles rendering of HTML templates.
	 It organizes templates into partials, layouts, and pages, and provides
	 methods for rendering them.

	 Example Usage:

		site := NewSite("/path/to/templates", template.FuncMap{"customFunc": myFunc})
		err := site.Render(writer, "page.layout", data)

	 This will render the template found in "/path/to/templates/page.layout.gohtml" using
	 the data provided, and write the output to the writer.
*/
type Site struct {
	t *template.Template
	w io.Writer

	partials map[string]string
	layouts  map[string]string
	pages    map[string]string
}

// NewSite initializes and returns a new Site object.
// rootPath: Directory where the templates are stored.
// funcMap: Template functions to be used.
func NewSite(rootPath string, funcMap template.FuncMap) *Site {
	return new(Site).parseTemplates(rootPath, funcMap)
}

func (s *Site) RenderPartialFunc(partialName string, data any) (string, error) {
	var buf bytes.Buffer

	// Read partial template from file
	t, err := template.ParseFiles(s.partials[partialName])
	if err != nil {
		return "", err
	}

	// Execute the partial template
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Render renders a template.
// w: Writer interface where output is written.
// viewName: Type of the template to render, may include layout e.g "page.layout".
// data: Data to be passed to the template.
func (s *Site) Render(w io.Writer, viewName string, data interface{}) error {
	viewNameParts := strings.Split(viewName, ".") // 0: Page; 1: Layout

	//var pageName string
	var layoutName string

	if len(viewNameParts) == 2 {
		//pageName = viewNameParts[0]
		layoutName = viewNameParts[1]
	} else if len(viewNameParts) == 1 {
		//pageName = viewNameParts[0]
		layoutName = "index"
	}

	//t := template.Must(s.t.Clone())
	//return template.Must(t.ParseFiles(append(s.partials, s.layouts[layoutName], s.pages[pageName])...)).ExecuteTemplate(w, layoutName, data)
	return s.t.ExecuteTemplate(w, layoutName, data)
}

// parseTemplates parses templates from rootPath and organizes them into
// layout, partial, and page categories. It also attaches any custom functions.
func (s *Site) parseTemplates(rootPath string, funcMap template.FuncMap) *Site {
	s.pages = make(map[string]string)
	s.layouts = make(map[string]string)

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
			s.pages[ext[0]] = path
		case "partial":
			s.partials[ext[0]] = path
		case "layout":
			s.layouts[ext[0]] = path

		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	s.t = template.New("").Funcs(funcMap)

	return s
}
