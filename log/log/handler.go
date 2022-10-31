package log

import (
	"github.com/gorilla/websocket"
	static "github.com/vaiktorg/grimoire/log/viewer/templates"

	"html/template"
	"log"
	"net/http"
)

type WebSocketViewer struct {
	template *template.Template
	client   *Client
	appName  string
}

func NewWebSocketViewer(AppName string) (*WebSocketViewer, error) {
	return &WebSocketViewer{
		template: template.Must(template.ParseFS(static.TemplateFS, "./*.gohtml")),
		appName:  AppName,
	}, nil
}

func (c *WebSocketViewer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := c.template.ExecuteTemplate(w, "index", nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
	m.HandleFunc("/ws", c.handler)
	m.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	m.ServeHTTP(w, r)
}

// WSHandler will ...
func (c *WebSocketViewer) handler(w http.ResponseWriter, r *http.Request) {
	if c.client == nil {
		http.Error(w, "client already connected", 300)
	}

	// Upgrade our raw HTTP connection to a websocket based one
	httpUpgrade := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := httpUpgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// The event loop
	c.client = NewClient(c.appName, conn)
}
