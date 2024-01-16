package simws

import (
	"context"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/uid"
	"github.com/vaiktorg/grimoire/util"
	"net/http"
	"nhooyr.io/websocket"
)

type IWebSocket interface {
	Dial(addr string, saveSess bool) (*Client, error)
	Broadcast(key string, data any)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	OnMessage(handler func(msg Message, sess *ConnSession))
}
type WebSocket struct {
	sess *SimSessions

	hooks     util.Hooks[*ConnSession]
	onMessage func(msg Message, sess *ConnSession)
	logger    log.ILogger
}

func NewWebSocket(config *Config) *WebSocket {
	sws := &WebSocket{
		sess:   NewSimConnSessions(),
		logger: config.Logger,
	}

	defer sws.logger.TRACE("NewWebSocket initialized")

	return sws
}

func (s *WebSocket) Dial(addr string, saveSess bool) (*Client, error) {
	conn, _, err := websocket.Dial(context.Background(), addr, nil)
	if err != nil {
		return nil, err
	}

	client, err := NewSimClient(conn)
	if err != nil {
		return nil, err
	}

	if saveSess {
		if err = s.sess.NewSession(client); err != nil {
			return nil, err
		}

		client.hooks.OnHook(util.OnDisconnect, func(client *Client) {
			err = s.sess.Disconnect(client.id)
			if err != nil {
				s.logger.ERROR(err.Error())
			}
		})
	}

	defer s.logger.TRACE("Dialed "+string(client.ID())+" was broadcast", client.id)

	return client, nil
}
func (s *WebSocket) Broadcast(key string, data any) {
	s.sess.sessions.Iterate(func(id uid.UID, sess *ConnSession) {
		sess.Client.Send(key, data)
	})

	s.logger.TRACE("message "+key+" was broadcast", data)
}
func (s *WebSocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {

		http.Error(w, s.logger.ERROR(err.Error()), http.StatusInternalServerError)
		return
	}

	client, err := NewSimClient(conn)
	if err != nil {
		http.Error(w, s.logger.ERROR(err.Error()), http.StatusInternalServerError)
		return
	}

	if err = s.sess.NewSession(client); err != nil {
		http.Error(w, s.logger.ERROR(err.Error()), http.StatusInternalServerError)
		return
	}

	go client.Listen(func(m Message) {
		sessId := s.sess.Session(client.id)
		s.logger.TRACE("onMessage event for: ", m, client.id, sessId)
		s.onMessage(m, sessId)
	})

	s.hooks.EnqueueHook(util.OnConnect)
	s.logger.TRACE("New Connection: " + string(client.id))
}

func (s *WebSocket) OnMessage(handler func(Message, *ConnSession)) {
	s.onMessage = handler
}
