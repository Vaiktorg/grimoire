package tmpl

import (
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

type Site struct {
	t *template.Template

	tmplPaths   []string
	layoutPaths map[string]string
	pagesPaths  map[string]string
}

func NewSite(rootPath string, funcMap template.FuncMap) echo.Renderer {
	return new(Site).parseTemplates(rootPath, funcMap)
}

// Render TODO: Test if Through render you can send HTML Blob and JSON after.
func (s *Site) Render(w io.Writer, viewName string, data interface{}, _ echo.Context) error {
	viewNameParts := strings.Split(viewName, ".") // 0: Page; 1: Layout

	var pageName string
	var layoutName string

	if len(viewNameParts) == 2 {
		pageName = viewNameParts[0]
		layoutName = viewNameParts[1]
	} else if len(viewNameParts) == 1 {
		pageName = viewNameParts[0]
		layoutName = "index"
	}

	t := template.Must(s.t.Clone())
	return template.Must(t.ParseFiles(append(s.tmplPaths, s.layoutPaths[layoutName], s.pagesPaths[pageName])...)).ExecuteTemplate(w, layoutName, data)
}

func (s *Site) parseTemplates(rootPath string, funcMap template.FuncMap) echo.Renderer {
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
			s.tmplPaths = append(s.tmplPaths, path)
		case "layout":
			s.layoutPaths[ext[0]] = path

		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	s.t = template.New("").Funcs(funcMap)

	return s
}

type ViewRender struct {
	Engine *Site
	Name   string
	Data   any
}

// Instance implement gin interface
func (s *Site) Instance(name string, data any) ViewRender {
	return ViewRender{
		Engine: s,
		Name:   name,
		Data:   data,
	}
}
