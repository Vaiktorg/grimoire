package simws

import (
	"context"
	"github.com/vaiktorg/grimoire/serve/ws"
	"github.com/vaiktorg/grimoire/uid"
	"net/http"
	"nhooyr.io/websocket"
)

type EventType string

const (
	OnConnect    EventType = "connect"
	OnMessage              = "message"
	OnDisconnect           = "disconnect"
)

func (e *EventType) Is(eventType EventType) bool {
	return *e == eventType
}

type ISimWebSocket interface {
	Dial(addr string, saveSess bool) (ws.IClient, error)
	Broadcast(key string, data any)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	OnEvent(handler func(evType EventType, sess *SimConnSession))
	OnMessage(handler func(msg ws.Message, sess *SimConnSession))
}
type SimWebSocket struct {
	sess *SimSessions
	ctx  context.Context

	eventChan chan func()

	onEvent   func(evType EventType, sess *SimConnSession)
	onMessage func(msg ws.Message, sess *SimConnSession)
}

func NewSimWebSocket(ctx context.Context) *SimWebSocket {
	sws := &SimWebSocket{
		sess:      NewSimConnSessions(),
		eventChan: make(chan func(), 10),
		ctx:       ctx,
	}

	go sws.listenForEvents()

	return sws
}

func (s *SimWebSocket) Dial(addr string, saveSess bool) (ws.IClient, error) {
	conn, _, err := websocket.Dial(s.ctx, addr, nil)
	if err != nil {
		return nil, err
	}

	client := NewSimClient(conn)
	if saveSess {
		if err = s.sess.NewSession(client, nil); err != nil {
			return nil, err
		}
	}

	return client, nil
}
func (s *SimWebSocket) Broadcast(key string, data any) {
	s.sess.sessions.Iterate(func(id uid.UID, sess *SimConnSession) {
		sess.Client.Send(key, data)
	})

}
func (s *SimWebSocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := NewSimClient(conn)
	if err = s.sess.NewSession(client, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client.OnEvent(func(eventType EventType, client ws.IClient) {
		if eventType.Is(OnDisconnect) {
			_ = s.sess.Disconnect(client.ID())
		}
	})

	go client.Listen(func(m ws.Message) {
		s.onMessage(m, s.sess.Session(client.id))
	})

	client.OnEvent(func(eventType EventType, client ws.IClient) {
		if eventType.Is(OnDisconnect) {
			s.enqueueEvents(eventType, s.sess.Session(client.ID()))
		}
	})
	s.enqueueEvents(OnConnect, s.sess.Session(client.id))
}

func (s *SimWebSocket) OnEvent(handler func(eventType EventType, sess *SimConnSession)) {
	s.onEvent = handler
}
func (s *SimWebSocket) OnMessage(handler func(message ws.Message, sess *SimConnSession)) {
	s.onMessage = handler
}

func (s *SimWebSocket) listenForEvents() {
	for event := range s.eventChan {
		event()
	}
}
func (s *SimWebSocket) enqueueEvents(evType EventType, sess *SimConnSession) {
	s.eventChan <- func() {
		if s.onEvent != nil {
			s.onEvent(evType, sess)
		}
	}
}
