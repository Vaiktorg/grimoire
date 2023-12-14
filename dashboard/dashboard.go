package dashboard

import (
	"bytes"
	"fmt"
	"github.com/vaiktorg/grimoire/log"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Dashboard struct {
	running bool
	mux     *http.ServeMux
	tmpl    *template.Template
	logger  log.ILogger
}

type Config struct {
	Logger log.ILogger
}

func NewDashboard(config Config) *Dashboard {
	d := &Dashboard{
		tmpl: template.Must(template.New("").Funcs(template.FuncMap{
			"lower": strings.ToLower,
		}).ParseGlob("templates/*.gohtml")),
		logger: config.Logger,
		mux:    http.NewServeMux(),
	}

	if config.Logger == nil {
		panic("logger is nil")
	}

	registerMux(d)

	d.logger.TRACE("New Dashboard Handler")

	return d
}

func registerMux(d *Dashboard) {
	d.mux.HandleFunc("/", d.homeHandler)

	// Static Content
	d.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Logs SSE Handler
	d.mux.HandleFunc("/sse", d.loggerHandler)

	// FormActions
	d.mux.HandleFunc("/actions", d.actionsMux)
}

func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.mux.ServeHTTP(w, r)
}

func (d *Dashboard) homeHandler(w http.ResponseWriter, _ *http.Request) {
	err := d.tmpl.ExecuteTemplate(w, "index", nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (d *Dashboard) loggerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	d.checkLastMessages(r)

	if d.logger != nil {
		d.logger.Output(func(logMsg log.Log) {
			if d.running {
				d.executeLogPartial(w, logMsg)
			}
		})
	}
}
func (d *Dashboard) checkLastMessages(r *http.Request) {
	// Check for last cached messages
	lastId := r.Header.Get("Last-Event-ID")
	if lastId == "" {
		return
	}

	logId, err := strconv.Atoi(lastId)
	if err != nil {
		return
	}

	// LogID are sequential N+1, meaning if last ClientID is 563,
	// then it should return the last 563 logs since runtime.
	cached := d.logger.Messages(log.Pagination{
		Page:   1,
		Amount: logId,
	})
	if d.logger.TotalSent() < uint64(logId) || cached == nil {
		return
	}

	d.logger.BatchLogs(cached...)
}
func (d *Dashboard) executeLogPartial(w http.ResponseWriter, msgs ...log.Log) {
	partialBuff := bytes.Buffer{}
	err := d.tmpl.ExecuteTemplate(&partialBuff, "src-card", msgs)
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
