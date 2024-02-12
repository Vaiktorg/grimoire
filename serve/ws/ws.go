package ws

import (
	"context"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/uid"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
)

type IConn interface {
	Read(ctx context.Context) (websocket.MessageType, []byte, error)
	Write(ctx context.Context, typ websocket.MessageType, data []byte) error
	Close(code websocket.StatusCode, reason string) error
}

type IWebSocket interface {
	Sessions() *Sessions
	Dial(string) (IClient, error)
	Broadcast(string, any)
	OnMessage(func(Message, *ConnSession))
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type WebSocket struct {
	mu sync.Mutex
	wg *sync.WaitGroup

	sess   *Sessions
	logger log.ILogger

	onMessage func(Message, *ConnSession)
}

func NewWebSocket(config *Config) *WebSocket {
	return &WebSocket{
		mu:     sync.Mutex{},
		wg:     new(sync.WaitGroup),
		sess:   config.Sessions,
		logger: config.Logger,
	}
}

func (ws *WebSocket) Dial(addr string) (IClient, error) {
	conn, _, err := websocket.Dial(context.Background(), addr, nil)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(conn)
	if err != nil {
		return nil, err
	}

	if ws.onMessage != nil {
		go func() {
			client.OnMessage(func(m Message) {
				sessId := ws.sess.Session(client.id)
				//ws.logger.TRACE("onMessage event for: ", m, client.id, sessId)
				ws.onMessage(m, sessId)
			})
		}()
	}

	return client, nil
}
func (ws *WebSocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client, err := NewClient(conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = ws.sess.NewSession(client); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Listen
	if ws.onMessage != nil {
		go func() {
			client.OnMessage(func(m Message) {
				sessId := ws.sess.Session(client.id)
				//ws.logger.TRACE("onMessage event for: ", m, client.id, sessId)
				ws.onMessage(m, sessId)
			})
		}()
	}
}

// OnMessage assigns a single handler where we receive a ws.Message straight from the ws.Client's proc loop.
func (ws *WebSocket) OnMessage(handler func(message Message, client *ConnSession)) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.onMessage = handler
}

// Broadcast sends a message to every connected Client
func (ws *WebSocket) Broadcast(key string, data any) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.sess.Iterate(func(uid uid.UID, session *ConnSession) {
		session.Client.Send(key, data)
	})
}
func (ws *WebSocket) Sessions() *Sessions {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	return ws.sess
}
