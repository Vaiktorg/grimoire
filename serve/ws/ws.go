package ws

import (
	"context"
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
	Dial(string, bool) (IClient, error)
	Broadcast(key string, data any)
	OnMessage(func(Message))
}

type WebSocket struct {
	mu  sync.Mutex
	wg  *sync.WaitGroup
	ctx context.Context

	sess *Sessions

	gob bool // false: json, true: gob

	onMessage    func(Message)
	onConnect    []func(client IClient)
	onDisconnect []func(client IClient)
}

func NewWebSocket() *WebSocket {
	return &WebSocket{
		wg:   new(sync.WaitGroup),
		sess: NewConnSessions(),
	}
}

func (s *WebSocket) Dial(addr string, saveSess bool) (IClient, error) {
	conn, _, err := websocket.Dial(s.ctx, addr, nil)
	if err != nil {
		return nil, err
	}

	client := NewClient(conn)

	if !saveSess {
		return client, nil
	}

	if err = s.sess.NewSession(client); err != nil {
		return nil, err
	}

	client.OnConnect(s.onConnect...)
	client.OnDisconnect(s.onDisconnect...)
	client.OnDisconnect(func(client IClient) {
		println("Client OnDisconnect " + client.ID())
		_ = s.sess.Disconnect(client.ID())
	})

	return client, nil
}
func (s *WebSocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := NewClient(conn)

	if err = s.sess.NewSession(client); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		_ = conn.Close(websocket.StatusInternalError, err.Error())
		return
	}

	client.OnConnect(s.onConnect...)
	client.OnDisconnect(s.onDisconnect...)
	client.OnDisconnect(func(client IClient) {
		println("Client OnDisconnect " + client.ID())
		_ = s.sess.Disconnect(client.ID())
	})

	// Listen
	if s.onMessage != nil {
		go client.OnMessage(s.onMessage)

		// Send Token
		client.Send("Token", s.sess.Session(client.id).Token)
	}
}

// OnMessage assigns a single handler where we receive a ws.Message straight from the ws.Client's proc loop.
func (s *WebSocket) OnMessage(handler func(message Message)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onMessage = handler
}

// OnConnect appends a handler that will be called at the end of the Client's connection sequence.
func (s *WebSocket) OnConnect(handler ...func(client IClient)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onConnect = append(s.onConnect, handler...)
}

// OnDisconnect appends a handler that will be called at the end of the Client's disconnection sequence.
func (s *WebSocket) OnDisconnect(handler ...func(client IClient)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onDisconnect = append(s.onDisconnect, handler...)
}

// Broadcast sends a message to every connected Client
func (s *WebSocket) Broadcast(key string, data any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sess.sessions.Iterate(func(k uid.UID, session *ConnSession) {
		session.client.Send(key, data)
	})
}

func (s *WebSocket) Sessions() *Sessions {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sess
}
