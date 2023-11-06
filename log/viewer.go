package log

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var WD string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	WD = wd
}

type Viewer struct {
	name      string
	t         *template.Template
	l         *Logger
	isRunning bool
	lastId    string
}

func NewLogViewer(appName string, logger *Logger) *Viewer {
	path := filepath.Join(WD, "/viewer/templates/*.gohtml")

	return &Viewer{
		t: template.Must(template.New("").Funcs(template.FuncMap{
			"lower": strings.ToLower,
		}).ParseGlob(path)),
		name: appName,
		l:    logger,
	}
}

func (c *Viewer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err = c.t.ExecuteTemplate(w, "index", nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	path := filepath.Join(wd, "/viewer/static")
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(path))))
	m.HandleFunc("/sse", c.SSEHandler)
	m.HandleFunc("/onoff", c.OnOffHandler)

	m.ServeHTTP(w, r)
}

func (c *Viewer) SSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	renderedPartial := bytes.Buffer{}

	//lastId := r.Header.Get("Last-Event-ID")

	for msg := range c.l.Output {
		if !c.isRunning {
			continue
		}

		if c.isRunning {
			err := c.t.ExecuteTemplate(&renderedPartial, "pkg-card", msg)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			writeSSE(w, msg.ID, renderedPartial)
		}
	}
}

func (c *Viewer) OnOffHandler(w http.ResponseWriter, r *http.Request) {
	b, err := strconv.ParseBool(r.FormValue("isRunning"))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	c.isRunning = b
}

func writeSSE(w http.ResponseWriter, id uint64, renderedPartial bytes.Buffer) {
	// Send the rendered row as an SSE event
	_, err := fmt.Fprintf(w, "id: %d\ntype:%s\ndata: %s\n\n", id, "message", strings.TrimSpace(renderedPartial.String()))
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	w.(http.Flusher).Flush()
}
