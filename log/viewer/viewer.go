package viewer

import (
	"bytes"
	"fmt"
	"github.com/vaiktorg/grimoire/log"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Viewer struct {
	tmpl    *template.Template
	logger  log.ILogger
	running bool
	mux     *http.ServeMux
}

type Config struct {
	Logger       log.ILogger
	TemplatePath string
}

func NewLogViewer(config Config) *Viewer {
	v := &Viewer{
		tmpl: template.Must(template.New("").Funcs(template.FuncMap{
			"lower": strings.ToLower,
		}).ParseGlob(config.TemplatePath + "/*.gohtml")),
		logger: config.Logger,
		mux:    http.NewServeMux(),
	}

	v.registerMux()

	return v
}

func (v *Viewer) registerMux() {
	v.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := v.tmpl.ExecuteTemplate(w, "index", nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})

	// Static Content
	v.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./viewer/static"))))

	// Logs SSE Handler
	v.mux.HandleFunc("/sse", v.SSEHandler)

	// FormActions
	v.mux.HandleFunc("/actions", v.ActionsMux)
}

func (v *Viewer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.mux.ServeHTTP(w, r)
}

func (v *Viewer) SSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	v.checkLastMessages(r)

	v.logger.Output(func(logMsg log.Log) error {
		if v.running {
			v.executeLogPartial(w, logMsg)
		}

		return nil
	})
}
func (v *Viewer) checkLastMessages(r *http.Request) {
	// Check for last cached messages
	lastId := r.Header.Get("Last-Event-ID")
	if lastId == "" {
		return
	}

	logId, err := strconv.Atoi(lastId)
	if err != nil {
		return
	}

	// LogID are sequential N+1, meaning if last ID is 563,
	// then it should return the last 563 logs since runtime.
	cached := v.logger.Messages(log.Pagination{
		Page:   1,
		Amount: logId,
	})
	if v.logger.TotalSent() < uint64(logId) || cached == nil {
		return
	}

	v.logger.BatchLogs(cached...)
}
func (v *Viewer) executeLogPartial(w http.ResponseWriter, msgs ...log.Log) {
	partialBuff := bytes.Buffer{}
	err := v.tmpl.ExecuteTemplate(&partialBuff, "pkg-card", msgs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	for _, msg := range msgs {
		// Send the rendered row as an SSE event
		_, err = fmt.Fprintf(w, "id: %d\ntype:%s\ndata: %v\n\n", msg.ID, "message", strings.TrimSpace(partialBuff.String()))
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.(http.Flusher).Flush()
	}
}
