package dashboard

import (
	"context"
	"fmt"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/serve/simws"
	"github.com/vaiktorg/grimoire/serve/ws"
	"github.com/vaiktorg/grimoire/uid"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

type Dashboard struct {
	running bool

	mux  *http.ServeMux
	tmpl *template.Template

	ws     simws.IWebSocket
	logger log.ILogger
}

type Config struct {
	Logger            log.ILogger
	TemplateDirectory string
	StaticDirectory   string
	Context           context.Context
}

func NewDashboard(config *Config) (*Dashboard, error) {
	// Template directory
	templateDir := config.TemplateDirectory
	if templateDir == "" {
		templateDir = "templates" // default path
	}

	// Parse templates
	tmpl, err := template.New("").Funcs(template.FuncMap{"lower": strings.ToLower}).ParseGlob(filepath.Join(templateDir, "*.gohtml"))
	if err != nil {
		return nil, fmt.Errorf("error parsing templates: %w", err)
	}

	// Static directory
	staticDir := config.StaticDirectory
	if staticDir == "" {
		staticDir = "./static" // default path
	}

	d := &Dashboard{
		mux: http.NewServeMux(),
		ws: simws.NewWebSocket(&ws.Config{
			GlobalID: uid.NewUID(64),
			Logger: config.Logger.NewServiceLogger(&log.Config{
				CanPrint:    true,
				CanOutput:   true,
				Persist:     true,
				ServiceName: "SimWebSocketConfig",
			}),
		}),
		tmpl:   tmpl,
		logger: config.Logger,
	}

	registerMux(d, staticDir)

	d.logger.TRACE("New Dashboard Handler")

	return d, nil
}

func registerMux(d *Dashboard, staticDir string) {
	d.mux.HandleFunc("/", d.homeHandler)
	d.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir)))) // Static Content

	d.mux.HandleFunc("/ws", d.ws.ServeHTTP)    // Logs SSE Handler
	d.mux.HandleFunc("/actions", d.actionsMux) // FormActions
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

func (d *Dashboard) ListenForClientEvents(host string) error {
	client, err := d.ws.Dial(host, true)
	if err != nil {
		return err
	}

	go client.Listen(func(msg simws.Message) {
		switch msg.KEY {
		case "log":
			client.Send(msg.KEY, msg.Data)
		}
	})

	return nil
}
