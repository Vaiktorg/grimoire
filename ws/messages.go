package ws

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"net/http"
	"strings"
	"sync"
)

type Events struct {
	msgs map[string]IMsg
	mu   sync.Mutex
}

func NewWSEvents() *Events {
	return &Events{
		msgs: make(map[string]IMsg),
	}
}
func (a *Events) Register(msg ...IMsg) {
	for _, iMsg := range msg {
		key := iMsg.Key()
		if _, ok := a.msgs[key]; !ok {
			a.msgs[key] = iMsg
		} else {
			log.Errorf("duplicate key found: %s", key)
		}
	}
}

func (a *Events) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrade := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		_, _ = fmt.Fprint(w, err)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		fmt.Println(string(msg))
		if err != nil {
			a.error(conn, err)
			continue
		}

		go func(c *websocket.Conn, aa *Events) {
			id, actionData, err := aa.parseMsg(msg)
			if err != nil {
				aa.error(conn, err)
				return
			}

			a.mu.Lock()
			defer a.mu.Unlock()
			action, ok := aa.msgs[id]
			if !ok {
				aa.error(conn, errors.New("could not find message for key: "+id))
				return
			}

			err = action.Process(json.NewDecoder(bytes.NewReader(actionData)), json.NewEncoder(c.UnderlyingConn()))
			if err != nil {
				aa.error(conn, err)
				return
			}
		}(conn, a)
	}
}

func (*Events) parseMsg(msg []byte) (string, []byte, error) {
	msg = []byte(strings.TrimPrefix(string(msg), "\""))

	data := bytes.SplitN(msg, []byte(":"), 2)
	if len(data) < 2 {
		return "", nil, errors.New("message does not contain key or data json")
	}

	if bytes.Compare(data[0], []byte("")) == 0 || len(data[0]) == 0 {
		return "", nil, errors.New("no id in socket message")
	}

	if !json.Valid(data[1]) {
		return "", nil, errors.New("no data in socket message")
	}

	return string(data[0]), data[1], nil
}
func (a *Events) error(conn *websocket.Conn, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err != nil {
		_ = conn.WriteJSON("error: " + err.Error())
	}
}

// --------------------------------------------------

// Message processes Decoding and Execution of IAction process types.
type Message[T IAction] struct {
	KEY  string
	data T
}

func (a *Message[T]) Key() string { return a.KEY }
func (a *Message[T]) Process(decoder *json.Decoder, ctx *json.Encoder) error {
	// Copy your message
	val := *a

	err := decoder.Decode(&val.data)
	if err != nil {
		return err
	}

	return val.data.Handler(ctx)
}

// --------------------------------------------------
