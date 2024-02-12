package docs

import (
	"bytes"
	"embed"
	"encoding/json"
	"github.com/vaiktorg/grimoire/markdown"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed index.gohtml style.css main.js
var f embed.FS

var JS template.JS
var CSS template.CSS

const ReadMeFileName = "readme.md"

// Wiki serves templates of rendered markdown files.
type Wiki struct {
	t        *template.Template
	md       *markdown.MDService
	isReadMe bool
	config   Config
	paths    map[string]string //K: docName, V: docPath
	toc      []string
}

type Config struct {
	Project  string `json:"project"`
	Desc     string `json:"desc"`
	DocsPath string `json:"docsPath"`
}

type Data struct {
	Title   string
	Article markdown.Article
	TOC     []string
	JS      template.JS
	CSS     template.CSS
}

// NewWiki is a constructor function that returns a pointer to a `Wiki` struct
func NewWiki(config Config) *Wiki {
	JS = template.JS(readFS("main.js"))
	CSS = template.CSS(readFS("style.css"))

	wk := &Wiki{
		t: template.Must(template.New("").Funcs(template.FuncMap{
			"base": func(query string) string {
				v, e := url.ParseRequestURI(query)
				if e != nil {
					return ""
				}
				return v.Query().Get("doc")
			},
		}).ParseFS(f, "*.gohtml")),
		md:     markdown.NewMarkdown(),
		config: config,
		paths:  make(map[string]string),
	}

	if err := wk.parseDocs(); err != nil {
		println(err)
		return wk
	}

	return wk
}

func (wk *Wiki) parseDocs() error {
	err := filepath.WalkDir(wk.config.DocsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		wk.paths[d.Name()] = path

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// A method that is called when the server receives a request.
func (wk *Wiki) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet: // Read some shit
		wk.GET(w, r)
	case http.MethodPost: // Upload some shit
		wk.POST(w, r)
	case http.MethodPatch: // Update some shit
		wk.PATCH(w, r)
	}
}

// GET Viewer for reading a document
func (wk *Wiki) GET(w http.ResponseWriter, r *http.Request) {
	docName := r.URL.Query().Get("doc")

	if docName == "" && wk.isReadMe {
		docName = ReadMeFileName
	}

	if !wk.Has(docName) {
		http.Error(w, "document "+docName+" not found.", http.StatusInternalServerError)
		return
	}

	fileP := filepath.Join(wk.config.DocsPath, docName)
	file, err := os.OpenFile(fileP, os.O_RDONLY, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	article, err := wk.md.Article(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = wk.t.ExecuteTemplate(w, "wiki", Data{
		Title:   wk.config.Project,
		Article: article,
		TOC:     wk.TOC(r.URL),
		JS:      JS,
		CSS:     CSS,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Has Checking if the file exists.
func (wk *Wiki) Has(docName string) bool {
	_, err := os.Stat(filepath.Join(wk.config.DocsPath, docName))
	if !os.IsExist(err) {
		return true
	}

	return false
}

// POST A method that is called when the server receives a POST request.
func (wk *Wiki) POST(w http.ResponseWriter, r *http.Request) {
	docName := r.URL.Query().Get("doc")

	if wk.Has(docName) {
		docName = strings.Join([]string{docName, time.Now().Format("2006-01-02")}, "_")
	}

	file, err := os.Create(filepath.Join(wk.config.DocsPath, docName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(file).Encode(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// PATCH A method that is called when the server receives a PATCH request.
func (wk *Wiki) PATCH(w http.ResponseWriter, r *http.Request) {
	var data bytes.Buffer

	docName := r.URL.Query().Get("doc")
	file, err := os.OpenFile(filepath.Join(wk.config.DocsPath, docName), os.O_RDWR, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func readFS(filename string) []byte {
	f, err := f.Open(filename)
	if err != nil {
		panic(err)
	}

	sts, err := f.Stat()
	if err != nil {
		panic(err)
	}

	buff := make([]byte, sts.Size())
	_, err = f.Read(buff)
	if err != nil {
		panic(err)
	}

	return buff
}

func (wk *Wiki) TOC(host *url.URL) []string {
	var toc []string

	for _, path := range wk.paths {
		name := filepath.Base(path)

		q := host.Query()
		q.Set("doc", name)
		host.RawQuery = q.Encode()
		if name == ReadMeFileName {
			toc = append([]string{host.RequestURI()}, toc...)
			wk.isReadMe = true
			continue
		}
		toc = append(toc, host.RequestURI())
	}

	return toc
}
